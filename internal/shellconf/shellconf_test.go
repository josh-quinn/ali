package shellconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/render"
)

func TestInstallAddsBlockOnceWithBackup(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	orig := "export A=1\nalias k=kubectl\n"
	os.WriteFile(rc, []byte(orig), 0o644)

	res, err := Install(rc)
	if err != nil || !res.Changed {
		t.Fatalf("first install: %+v %v", res, err)
	}
	got, _ := os.ReadFile(rc)
	if !strings.HasPrefix(string(got), orig) || !strings.Contains(string(got), render.BlockStart) {
		t.Errorf("original not preserved or block missing:\n%s", got)
	}
	bk, _ := os.ReadFile(res.Backup)
	if string(bk) != orig {
		t.Errorf("backup differs: %q", bk)
	}

	res2, err := Install(rc)
	if err != nil || res2.Changed || res2.Backup != "" {
		t.Errorf("second install should be no-op: %+v %v", res2, err)
	}
	again, _ := os.ReadFile(rc)
	if string(again) != string(got) {
		t.Error("idempotency violated")
	}
}

func TestInstallCreatesMissingFile(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".zshrc")
	res, err := Install(rc)
	if err != nil || !res.Changed || res.Backup != "" {
		t.Fatalf("%+v %v", res, err)
	}
	got, _ := os.ReadFile(rc)
	if !strings.Contains(string(got), render.BlockStart) {
		t.Error("block missing")
	}
}

func TestInstallEnsuresNewlineBeforeBlock(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".zshrc")
	os.WriteFile(rc, []byte("export A=1"), 0o644)
	Install(rc)
	got, _ := os.ReadFile(rc)
	if !strings.Contains(string(got), "export A=1\n") {
		t.Errorf("no newline separation: %q", got)
	}
}

func TestPathHonoursZDOTDIR(t *testing.T) {
	t.Setenv("ZDOTDIR", "/z")
	if p, err := Path(); err != nil || p != "/z/.zshrc" {
		t.Errorf("got %s %v", p, err)
	}
	t.Setenv("ZDOTDIR", "")
	t.Setenv("HOME", "/h")
	if p, err := Path(); err != nil || p != "/h/.zshrc" {
		t.Errorf("got %s %v", p, err)
	}
}

func TestPathNeverReturnsARelativePath(t *testing.T) {
	t.Setenv("ZDOTDIR", "")
	t.Setenv("HOME", "")
	if p, err := Path(); err == nil {
		t.Errorf("want error without HOME, got %q", p)
	}
	t.Setenv("ZDOTDIR", "relative")
	t.Setenv("HOME", "/h")
	if p, _ := Path(); p != "/h/.zshrc" {
		t.Errorf("relative ZDOTDIR not ignored: %q", p)
	}
}

func TestBackupOf(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".zshrc")
	os.WriteFile(rc, []byte("x"), 0o600)
	b, err := Backup(rc)
	if err != nil || !strings.Contains(filepath.Base(b), ".zshrc.ali-backup-") {
		t.Fatalf("%s %v", b, err)
	}
	if st, _ := os.Stat(b); st.Mode().Perm() != 0o600 {
		t.Errorf("backup mode %v", st.Mode())
	}
}

// Dotfile managers (stow, chezmoi, yadm) symlink .zshrc into a repo.
func TestInstallKeepsSymlinkedZshrc(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dotfiles", "zshrc")
	os.MkdirAll(filepath.Dir(target), 0o755)
	os.WriteFile(target, []byte("export A=1\n"), 0o644)
	rc := filepath.Join(dir, ".zshrc")
	os.Symlink(target, rc)

	res, err := Install(rc)
	if err != nil || !res.Changed {
		t.Fatalf("%+v %v", res, err)
	}
	if st, _ := os.Lstat(rc); st.Mode()&os.ModeSymlink == 0 {
		t.Error(".zshrc symlink was replaced by a regular file")
	}
	if b, _ := os.ReadFile(target); !strings.Contains(string(b), render.BlockStart) {
		t.Errorf("block not written into the symlink target:\n%s", b)
	}
}
