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

const oldBlock = render.BlockStart + "\nsome old ali code\n" + render.BlockEnd + "\n"

func TestInstallUpgradesAnOutdatedBlockInPlace(t *testing.T) {
	rc := filepath.Join(t.TempDir(), ".zshrc")
	before, after := "export A=1\nsource $ZSH/oh-my-zsh.sh\n\n", "\nalias late=x\n# trailing, no newline"
	os.WriteFile(rc, []byte(before+oldBlock+after), 0o600)

	res, err := Install(rc)
	if err != nil || !res.Changed || !res.Updated || res.Backup == "" {
		t.Fatalf("%+v %v", res, err)
	}
	got, _ := os.ReadFile(rc)
	if want := before + render.Block() + after; string(got) != want {
		t.Errorf("content outside the block changed:\n got %q\nwant %q", got, want)
	}
	if st, _ := os.Stat(rc); st.Mode().Perm() != 0o600 {
		t.Errorf("mode changed to %v", st.Mode().Perm())
	}
	if bk, _ := os.ReadFile(res.Backup); string(bk) != before+oldBlock+after {
		t.Error("backup is not the original file")
	}
	if res2, _ := Install(rc); res2.Changed {
		t.Error("second run should be a no-op")
	}
}

func TestInstallRefusesDamagedBlocks(t *testing.T) {
	cases := map[string]string{
		"no end":       "x\n" + render.BlockStart + "\nstuff\n",
		"no start":     "x\n" + render.BlockEnd + "\n",
		"end first":    render.BlockEnd + "\n" + render.BlockStart + "\n",
		"two blocks":   oldBlock + oldBlock,
		"nested start": render.BlockStart + "\n" + oldBlock,
	}
	for name, body := range cases {
		rc := filepath.Join(t.TempDir(), ".zshrc")
		os.WriteFile(rc, []byte(body), 0o644)
		if _, err := Install(rc); err == nil {
			t.Errorf("%s: want error", name)
		}
		if got, _ := os.ReadFile(rc); string(got) != body {
			t.Errorf("%s: file was modified", name)
		}
		if m, _ := filepath.Glob(rc + ".ali-backup-*"); len(m) != 0 {
			t.Errorf("%s: backup written for a refused change", name)
		}
	}
}

func TestInspect(t *testing.T) {
	cases := map[string]BlockState{
		"":                          BlockMissing,
		"x\n" + render.Block():      BlockCurrent,
		"x\n" + oldBlock:            BlockOutdated,
		render.BlockStart + "\nx\n": BlockDamaged,
	}
	for body, want := range cases {
		if got, _, _ := Inspect(body); got != want {
			t.Errorf("Inspect(%q) = %v, want %v", body, got, want)
		}
	}
	_, start, _ := Inspect("a\nb\n" + render.Block())
	if start != 3 {
		t.Errorf("block start line = %d, want 3", start)
	}
}
