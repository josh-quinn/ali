package store

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/josh-quinn/ali/internal/alias"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return New(t.TempDir())
}

func TestMissingFileIsEmpty(t *testing.T) {
	s := newStore(t)
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.All()) != 0 {
		t.Errorf("want empty, got %v", d.All())
	}
}

func TestRoundTrip(t *testing.T) {
	s := newStore(t)
	err := s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "git"}, false)
		d.Set(alias.Alias{Name: "hgp", Command: "helm get pods", Group: "k8s"}, true)
		d.Set(alias.Alias{Name: "q", Command: "echo 'it'\\''s' \"$X\"\nnext", Group: "echo"}, false)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	a, ok := d.Get("q")
	if !ok || a.Command != "echo 'it'\\''s' \"$X\"\nnext" {
		t.Errorf("special-char command lost: %+v", a)
	}
	if !d.Pinned("hgp") || d.Pinned("gp") {
		t.Error("pinned flags wrong")
	}
	if g, _ := d.Get("hgp"); g.Group != "k8s" {
		t.Errorf("group = %q", g.Group)
	}
	z, err := os.ReadFile(filepath.Join(s.Dir, "aliases.zsh"))
	if err != nil || !strings.Contains(string(z), "alias -- gp='git pull'") {
		t.Errorf("aliases.zsh not rendered: %v %s", err, z)
	}
}

func TestDeleteDropsEmptyGroup(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "git"}, false)
		return nil
	})
	s.Update(func(d *Data) error { d.Delete("gp"); return nil })
	raw, _ := os.ReadFile(s.File())
	if strings.Contains(string(raw), "[git]") {
		t.Errorf("empty group table kept:\n%s", raw)
	}
}

func TestMoveBetweenGroupsKeepsSingleCopy(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "git"}, false)
		return nil
	})
	s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "tools"}, true)
		return nil
	})
	d, _ := s.Load()
	if len(d.All()) != 1 {
		t.Errorf("want 1 alias, got %v", d.All())
	}
}

func TestHandEditedFileIsHonoured(t *testing.T) {
	s := newStore(t)
	os.WriteFile(s.File(), []byte("version = 1\n\n[git]\ngs = \"git status\"\n"), 0o600)
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if a, ok := d.Get("gs"); !ok || a.Group != "git" {
		t.Errorf("hand-edited alias not loaded: %+v", a)
	}
}

func TestInvalidFilesError(t *testing.T) {
	cases := map[string]string{
		"syntax":    "version = 1\n[git\n",
		"duplicate": "version = 1\n[a]\nx = \"1\"\n[b]\nx = \"2\"\n",
		"badname":   "version = 1\n[a]\n\"bad name\" = \"1\"\n",
		"nonstring": "version = 1\n[a]\nx = 3\n",
		"version":   "version = 99\n",
		"toplevel":  "version = 1\nstray = \"x\"\n",
	}
	for name, body := range cases {
		s := newStore(t)
		os.WriteFile(s.File(), []byte(body), 0o600)
		if _, err := s.Load(); err == nil {
			t.Errorf("%s: want error", name)
		} else if !strings.Contains(err.Error(), "aliases.toml") {
			t.Errorf("%s: error should name the file: %v", name, err)
		}
		// a failed update must not overwrite the broken file
		before, _ := os.ReadFile(s.File())
		s.Update(func(d *Data) error { return nil })
		after, _ := os.ReadFile(s.File())
		if string(before) != string(after) {
			t.Errorf("%s: invalid store was overwritten", name)
		}
	}
}

func TestUpdateErrorLeavesStoreUntouched(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error { d.Set(alias.Alias{Name: "a", Command: "x", Group: "g"}, false); return nil })
	before, _ := os.ReadFile(s.File())
	err := s.Update(func(d *Data) error { d.Delete("a"); return os.ErrInvalid })
	if err == nil {
		t.Fatal("want error")
	}
	after, _ := os.ReadFile(s.File())
	if string(before) != string(after) {
		t.Error("store changed despite error")
	}
}

func TestConcurrentUpdatesLoseNothing(t *testing.T) {
	s := newStore(t)
	var wg sync.WaitGroup
	names := []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "b1", "b2", "b3"}
	for _, n := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Update(func(d *Data) error {
				d.Set(alias.Alias{Name: n, Command: "echo " + n, Group: "echo"}, false)
				return nil
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	d, _ := s.Load()
	if len(d.All()) != len(names) {
		t.Errorf("lost updates: got %d of %d", len(d.All()), len(names))
	}
}

func TestDirResolution(t *testing.T) {
	get := func() string {
		t.Helper()
		d, err := DefaultDir()
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	t.Setenv("ALI_CONFIG_DIR", "/x/ali")
	if got := get(); got != "/x/ali" {
		t.Errorf("ALI_CONFIG_DIR: %s", got)
	}
	t.Setenv("ALI_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	if got := get(); got != "/xdg/ali" {
		t.Errorf("XDG: %s", got)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/h")
	if got := get(); got != "/h/.config/ali" {
		t.Errorf("HOME: %s", got)
	}
}

// With no home the old code produced the relative path ".config/ali" and wrote
// into whatever directory the process happened to be in.
func TestDirResolutionNeverFallsBackToTheWorkingDirectory(t *testing.T) {
	t.Setenv("ALI_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	if d, err := DefaultDir(); err == nil {
		t.Errorf("want error without a home directory, got %q", d)
	}
	t.Setenv("XDG_CONFIG_HOME", "relative/path") // the XDG spec says to ignore relative values
	t.Setenv("HOME", "/h")
	if d, _ := DefaultDir(); d != "/h/.config/ali" {
		t.Errorf("relative XDG_CONFIG_HOME not ignored: %q", d)
	}
}

func TestStoreFilesArePrivate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	s := New(dir)
	if err := s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "deploy", Command: "TOKEN=abc123 curl https://x", Group: "curl"}, false)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	mode := func(p string) os.FileMode { st, _ := os.Stat(p); return st.Mode().Perm() }
	if m := mode(dir); m != 0o700 {
		t.Errorf("config dir mode %v, want 0700", m)
	}
	for _, f := range []string{s.File(), s.ZshFile(), filepath.Join(dir, lockName)} {
		if m := mode(f); m != 0o600 {
			t.Errorf("%s mode %v, want 0600", filepath.Base(f), m)
		}
	}
}

func TestExistingLooseFilesAreTightenedButSharedDirIsLeftAlone(t *testing.T) {
	dir := t.TempDir() // already exists, may be shared: must not be chmod'ed
	os.Chmod(dir, 0o755)
	s := New(dir)
	os.WriteFile(s.File(), []byte("version = 1\n"), 0o644)
	if err := s.Update(func(d *Data) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Stat(s.File()); st.Mode().Perm() != 0o600 {
		t.Errorf("store mode %v, want 0600", st.Mode().Perm())
	}
	if st, _ := os.Stat(dir); st.Mode().Perm() != 0o755 {
		t.Errorf("pre-existing dir mode changed to %v", st.Mode().Perm())
	}
}

func TestWriteAtomicWritesThroughSymlinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dotfiles", "aliases.toml")
	os.MkdirAll(filepath.Dir(target), 0o755)
	os.WriteFile(target, []byte("old"), 0o640)
	link := filepath.Join(dir, "link.toml")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(link, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Lstat(link); st.Mode()&os.ModeSymlink == 0 {
		t.Error("symlink was replaced by a regular file")
	}
	if b, _ := os.ReadFile(target); string(b) != "new" {
		t.Errorf("target not updated: %q", b)
	}
}

func TestWriteAtomicCreatesThroughDanglingSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dotfiles", "aliases.toml")
	os.MkdirAll(filepath.Dir(target), 0o755)
	link := filepath.Join(dir, "link.toml")
	os.Symlink(filepath.Join("dotfiles", "aliases.toml"), link) // relative, target missing
	if err := WriteAtomic(link, []byte("fresh"), 0o600); err != nil {
		t.Fatal(err)
	}
	if st, _ := os.Lstat(link); st.Mode()&os.ModeSymlink == 0 {
		t.Error("symlink was replaced")
	}
	if b, _ := os.ReadFile(target); string(b) != "fresh" {
		t.Errorf("target not created: %q", b)
	}
}
