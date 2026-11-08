// Package store persists aliases in a hand-editable TOML file and keeps the
// derived aliases.zsh in sync. Writes are atomic and serialised by a file lock.
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/pelletier/go-toml/v2"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/render"
)

const (
	fileName    = "aliases.toml"
	zshName     = "aliases.zsh"
	lockName    = "aliases.toml.lock"
	fileVersion = 1
)

// Store is a handle on a config directory.
type Store struct{ Dir string }

// New returns a Store rooted at dir.
func New(dir string) *Store { return &Store{Dir: dir} }

// DefaultDir resolves the config directory: ALI_CONFIG_DIR, then
// $XDG_CONFIG_HOME/ali, then ~/.config/ali.
func DefaultDir() string {
	if d := os.Getenv("ALI_CONFIG_DIR"); d != "" {
		return d
	}
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "ali")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "ali")
}

// File is the path of the TOML store.
func (s *Store) File() string { return filepath.Join(s.Dir, fileName) }

// ZshFile is the path of the generated zsh file.
func (s *Store) ZshFile() string { return filepath.Join(s.Dir, zshName) }

// Data is the in-memory content of the store.
type Data struct {
	aliases map[string]alias.Alias
	pinned  map[string]bool
}

func newData() *Data {
	return &Data{aliases: map[string]alias.Alias{}, pinned: map[string]bool{}}
}

// Get returns the alias called name.
func (d *Data) Get(name string) (alias.Alias, bool) { a, ok := d.aliases[name]; return a, ok }

// Pinned reports whether name's group was set explicitly.
func (d *Data) Pinned(name string) bool { return d.pinned[name] }

// Set adds or replaces an alias; pinned records an explicit group choice.
func (d *Data) Set(a alias.Alias, pinned bool) {
	d.aliases[a.Name] = a
	if pinned {
		d.pinned[a.Name] = true
	} else {
		delete(d.pinned, a.Name)
	}
}

// Delete removes name and reports whether it existed.
func (d *Data) Delete(name string) bool {
	_, ok := d.aliases[name]
	delete(d.aliases, name)
	delete(d.pinned, name)
	return ok
}

// All returns every alias sorted by group, then name.
func (d *Data) All() []alias.Alias {
	out := make([]alias.Alias, 0, len(d.aliases))
	for _, a := range d.aliases {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Group != out[j].Group {
			return out[i].Group < out[j].Group
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// Load reads the store. A missing file is an empty store; a broken file is an error.
func (s *Store) Load() (*Data, error) {
	raw, err := os.ReadFile(s.File())
	if os.IsNotExist(err) {
		return newData(), nil
	}
	if err != nil {
		return nil, err
	}
	return parse(s.File(), raw)
}

func parse(path string, raw []byte) (*Data, error) {
	var top map[string]any
	if err := toml.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	d := newData()
	if v, ok := top["version"]; ok {
		if n, ok := v.(int64); !ok || n != fileVersion {
			return nil, fmt.Errorf("%s: unsupported version %v (this ali supports %d)", path, v, fileVersion)
		}
	}
	for key, val := range top {
		switch key {
		case "version":
		case "_pinned":
			t, _ := val.(map[string]any)
			names, _ := t["names"].([]any)
			for _, n := range names {
				if str, ok := n.(string); ok {
					d.pinned[str] = true
				}
			}
		default:
			table, ok := val.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s: unexpected top-level key %q (aliases belong in a [group] table)", path, key)
			}
			if err := alias.ValidateGroup(key); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
			for name, cv := range table {
				cmd, ok := cv.(string)
				if !ok {
					return nil, fmt.Errorf("%s: [%s] %s must be a string", path, key, name)
				}
				if err := alias.ValidateName(name); err != nil {
					return nil, fmt.Errorf("%s: [%s]: %w", path, key, err)
				}
				if prev, dup := d.aliases[name]; dup {
					return nil, fmt.Errorf("%s: alias %q appears in both [%s] and [%s]", path, name, prev.Group, key)
				}
				d.aliases[name] = alias.Alias{Name: name, Command: cmd, Group: key}
			}
		}
	}
	for n := range d.pinned {
		if _, ok := d.aliases[n]; !ok {
			delete(d.pinned, n)
		}
	}
	return d, nil
}

func (d *Data) marshal() ([]byte, error) {
	top := map[string]any{"version": int64(fileVersion)}
	for _, a := range d.aliases {
		t, _ := top[a.Group].(map[string]any)
		if t == nil {
			t = map[string]any{}
			top[a.Group] = t
		}
		t[a.Name] = a.Command
	}
	if len(d.pinned) > 0 {
		names := make([]string, 0, len(d.pinned))
		for n := range d.pinned {
			names = append(names, n)
		}
		sort.Strings(names)
		top["_pinned"] = map[string]any{"names": names}
	}
	return toml.Marshal(top)
}

// Update runs fn on the current data under an exclusive lock and, if fn
// succeeds, atomically saves the store and regenerates aliases.zsh. If the
// store file is invalid or fn fails, nothing on disk changes.
func (s *Store) Update(fn func(*Data) error) error {
	// Aliases may embed tokens, so a directory we create is private. A directory
	// that already exists is left alone: it might be shared (ALI_CONFIG_DIR).
	if _, err := os.Stat(s.Dir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.Dir, 0o700); err != nil {
			return err
		}
	}
	lock, err := os.OpenFile(filepath.Join(s.Dir, lockName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	d, err := s.Load()
	if err != nil {
		return err
	}
	if err := fn(d); err != nil {
		return err
	}
	body, err := d.marshal()
	if err != nil {
		return err
	}
	if err := WriteAtomic(s.File(), body, 0o600); err != nil {
		return err
	}
	return WriteAtomic(s.ZshFile(), []byte(render.Zsh(d.All())), 0o600)
}

// resolve follows symlinks so a symlinked file (stow, chezmoi, a dotfiles repo)
// is updated in place instead of being replaced by a regular file. A dangling
// symlink resolves to where it points, so the target gets created.
func resolve(path string) string {
	for i := 0; i < 16; i++ {
		fi, err := os.Lstat(path)
		if err != nil || fi.Mode()&os.ModeSymlink == 0 {
			return path
		}
		next, err := os.Readlink(path)
		if err != nil {
			return path
		}
		if !filepath.IsAbs(next) {
			next = filepath.Join(filepath.Dir(path), next)
		}
		path = next
	}
	return path
}

// WriteAtomic writes data to path via a temp file in the same directory,
// fsyncs, then renames, so readers never see a partial file. Symlinks are
// followed (see resolve).
func WriteAtomic(path string, data []byte, perm os.FileMode) error {
	path = resolve(path)
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
