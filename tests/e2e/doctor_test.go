package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// omzStub stands in for Oh My Zsh: a file sourced as $ZSH/oh-my-zsh.sh that
// defines plugin aliases.
func omzStub(t *testing.T, s *sandbox) string {
	dir := filepath.Join(s.home, "omz")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "oh-my-zsh.sh"), []byte(omzRC), 0o644)
	return "export ZSH=" + dir + "\nsource $ZSH/oh-my-zsh.sh\n"
}

func TestDoctorHealthyWithOhMyZsh(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte(omzStub(t, s)), 0o644)
	s.ali("init", "zsh")
	s.ali("add", "gp", "echo pulled")
	got := s.zsh("ali doctor; echo exit=$?")
	if strings.Contains(got, "FAIL") || !strings.Contains(got, "exit=0") || !strings.Contains(got, "ok      aliases not redefined") {
		t.Errorf("healthy setup reported problems:\n%s", got)
	}
}

func TestDoctorFindsRedefinitionAfterTheBlock(t *testing.T) {
	s := newSandbox(t)
	rc := filepath.Join(s.home, ".zshrc")
	os.WriteFile(rc, []byte(omzStub(t, s)), 0o644)
	s.ali("init", "zsh")
	s.ali("add", "gp", "echo pulled")
	b, _ := os.ReadFile(rc)
	os.WriteFile(rc, append(b, []byte("alias gp='git push --force'\n")...), 0o644)

	// run another ali command first: the check must still see the startup state
	got := s.zsh("ali ls >/dev/null\nali doctor; echo exit=$?")
	if !strings.Contains(got, "gp (now: git push --force)") || !strings.Contains(got, "exit=1") {
		t.Errorf("redefinition not reported:\n%s", got)
	}
}

func TestDoctorFindsBlockBeforeOhMyZsh(t *testing.T) {
	s := newSandbox(t)
	rc := filepath.Join(s.home, ".zshrc")
	s.ali("init", "zsh")
	b, _ := os.ReadFile(rc)
	os.WriteFile(rc, append(b, []byte(omzStub(t, s))...), 0o644)
	got := s.zsh("ali doctor; echo exit=$?")
	if !strings.Contains(got, "FAIL    block loads after Oh My Zsh") || !strings.Contains(got, "exit=1") {
		t.Errorf("order problem not reported:\n%s", got)
	}
}

func TestDoctorFromAScriptSkipsLiveCheck(t *testing.T) {
	s := newSandbox(t)
	s.ali("init", "zsh")
	out := s.ali("doctor")
	if !strings.Contains(out, "skipped aliases not redefined") {
		t.Errorf("output:\n%s", out)
	}
}
