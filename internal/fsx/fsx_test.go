package fsx

import (
	"os"
	"path/filepath"
	"testing"
)

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
