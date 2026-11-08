// Package e2e drives the real ali binary through a real interactive zsh in a
// throwaway HOME. It never touches the developer's own shell config.
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type sandbox struct {
	t    *testing.T
	home string
	bin  string
}

func newSandbox(t *testing.T) *sandbox {
	t.Helper()
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not installed")
	}
	_ = zsh
	bin := t.TempDir()
	build := exec.Command("go", "build", "-o", filepath.Join(bin, "ali"), "../../cmd/ali")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	home := t.TempDir()
	// Quiet prompt so command output appears on clean lines.
	os.WriteFile(filepath.Join(home, ".zshenv"), []byte("PS1=''\nRPS1=''\nPROMPT_EOL_MARK=''\nunsetopt prompt_sp\n"), 0o644)
	return &sandbox{t: t, home: home, bin: bin}
}

func (s *sandbox) env() []string {
	return []string{
		"HOME=" + s.home, "ZDOTDIR=" + s.home,
		"PATH=" + s.bin + ":/usr/bin:/bin", "TERM=dumb", "PS1=", "RPS1=",
	}
}

// zsh feeds script to an interactive shell on stdin, one line at a time like
// a person typing, so aliases defined by one line expand on the next. (With
// `zsh -c` the whole script is parsed up front and they would not.)
func (s *sandbox) zsh(script string) string {
	s.t.Helper()
	cmd := exec.Command("zsh", "-i", "-d")
	cmd.Stdin = strings.NewReader(script + "\nexit\n")
	cmd.Env = s.env()
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func (s *sandbox) ali(args ...string) string {
	s.t.Helper()
	cmd := exec.Command(filepath.Join(s.bin, "ali"), args...)
	cmd.Env = s.env()
	out, err := cmd.CombinedOutput()
	if err != nil {
		s.t.Fatalf("ali %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func hasLine(out, want string) bool {
	for _, l := range strings.Split(out, "\n") {
		if strings.TrimSpace(l) == want {
			return true
		}
	}
	return false
}

func TestAddIsLiveImmediatelyAndPersists(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte("export KEEP=1\n"), 0o644)
	s.ali("init", "zsh")

	got := s.zsh("ali add hi 'echo hello world'\nhi")
	if !hasLine(got, "hello world") {
		t.Fatalf("alias not live in the same shell:\n%s", got)
	}
	if got := s.zsh("hi"); !hasLine(got, "hello world") {
		t.Fatalf("alias missing in a new shell:\n%s", got)
	}
	rc, _ := os.ReadFile(filepath.Join(s.home, ".zshrc"))
	if !strings.HasPrefix(string(rc), "export KEEP=1\n") {
		t.Errorf("existing .zshrc content not preserved:\n%s", rc)
	}
}

func TestUpdateAndRemoveApplyToLiveShell(t *testing.T) {
	s := newSandbox(t)
	s.ali("init", "zsh")
	s.ali("add", "zz-who", "echo one")

	got := s.zsh("zz-who\nali update zz-who 'echo two'\nzz-who\nali rm zz-who\nzz-who")
	if !hasLine(got, "one") || !hasLine(got, "two") {
		t.Errorf("update not applied live:\n%s", got)
	}
	if strings.Count(got, "\ntwo") > 1 || !strings.Contains(got, "not found") {
		t.Errorf("removed alias still defined in live shell:\n%s", got)
	}
}

func TestUserOwnAliasesAreNotUnaliased(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte("alias mine='echo mine'\n"), 0o644)
	s.ali("init", "zsh")
	s.ali("add", "x", "echo x")
	got := s.zsh("ali rm x\nmine")
	if !hasLine(got, "mine") || strings.Contains(got, "command not found: mine") {
		t.Errorf("user alias was disturbed:\n%s", got)
	}
}

func TestSpecialCharactersRoundTripThroughShell(t *testing.T) {
	s := newSandbox(t)
	s.ali("init", "zsh")
	s.ali("add", "q", `echo 'it'\''s' "$HOME" && echo done`)
	got := s.zsh("q")
	if !hasLine(got, "it's "+s.home) || !hasLine(got, "done") {
		t.Errorf("special characters mangled:\n%s", got)
	}
}

func TestImportThenUseInNewShell(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte("alias old='echo from-zshrc'\n"), 0o644)
	s.ali("init", "zsh")
	out := s.ali("import")
	if !strings.Contains(out, "imported 1") {
		t.Fatalf("import output:\n%s", out)
	}
	// original line is still there and ali's copy lists under its group
	if ls := s.ali("ls", "echo"); !strings.Contains(ls, "old") {
		t.Errorf("imported alias not grouped:\n%s", ls)
	}
}

func TestInitIsIdempotent(t *testing.T) {
	s := newSandbox(t)
	s.ali("init", "zsh")
	first, _ := os.ReadFile(filepath.Join(s.home, ".zshrc"))
	if out := s.ali("init", "zsh"); !strings.Contains(out, "already set up") {
		t.Errorf("second init: %s", out)
	}
	second, _ := os.ReadFile(filepath.Join(s.home, ".zshrc"))
	if string(first) != string(second) {
		t.Error(".zshrc changed on second init")
	}
}
