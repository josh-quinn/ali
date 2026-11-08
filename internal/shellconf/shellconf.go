// Package shellconf edits the user's .zshrc: only the managed ali block, only
// on request, always after a timestamped backup.
package shellconf

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/josh-quinn/ali/internal/render"
	"github.com/josh-quinn/ali/internal/store"
)

// Path returns the .zshrc location, honouring ZDOTDIR.
func Path() string {
	if z := os.Getenv("ZDOTDIR"); z != "" {
		return filepath.Join(z, ".zshrc")
	}
	return filepath.Join(os.Getenv("HOME"), ".zshrc")
}

// Result describes what Install did.
type Result struct {
	Changed bool
	Backup  string // empty when no backup was needed
}

// HasBlock reports whether content already contains the ali block.
func HasBlock(content string) bool { return strings.Contains(content, render.BlockStart) }

// Backup copies path to <path>.ali-backup-<timestamp>, preserving its mode.
func Backup(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	dst := path + ".ali-backup-" + time.Now().Format("20060102-150405.000000000")
	return dst, store.WriteAtomic(dst, data, st.Mode().Perm())
}

// Install appends the ali block to the rc file unless it is already there.
func Install(path string) (Result, error) {
	cur, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}
	if HasBlock(string(cur)) {
		return Result{}, nil
	}
	var res Result
	mode := os.FileMode(0o644)
	if err == nil {
		if res.Backup, err = Backup(path); err != nil {
			return Result{}, err
		}
		if st, serr := os.Stat(path); serr == nil {
			mode = st.Mode().Perm()
		}
	}
	body := string(cur)
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	if body != "" {
		body += "\n"
	}
	body += render.Block()
	if err := store.WriteAtomic(path, []byte(body), mode); err != nil {
		return Result{}, err
	}
	res.Changed = true
	return res, nil
}
