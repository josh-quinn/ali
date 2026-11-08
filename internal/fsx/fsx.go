// Package fsx holds the one file-writing primitive every ali package uses:
// atomic, symlink-following writes (ADRs 0014, 0023).
package fsx

import (
	"os"
	"path/filepath"
)

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
	// After a successful rename the temp name no longer exists; ignore that error.
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close() // the write error is the one worth reporting
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close() // the write error is the one worth reporting
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close() // the write error is the one worth reporting
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
