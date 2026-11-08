package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// omzRC imitates what Oh My Zsh leaves in a shell before the ali block runs:
// plugin aliases (the git plugin's gp) and functions.
const omzRC = `alias gp='git push'
alias gst='git status'
mk() { echo "omz-mk $*"; }
`

func TestOverridingOhMyZshAliasesWarnsAndStillWorks(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte(omzRC), 0o644)
	s.ali("init", "zsh")

	got := s.zsh("ali add gp 'echo pulled'\ngp\nali add mk 'echo ali-mk'\nmk")
	if !strings.Contains(got, "warning: gp is already an alias in this shell (git push)") {
		t.Errorf("no override warning for gp:\n%s", got)
	}
	if !strings.Contains(got, "warning: mk is a shell function") {
		t.Errorf("no function warning for mk:\n%s", got)
	}
	if !hasLine(got, "pulled") || !hasLine(got, "ali-mk") {
		t.Errorf("ali's aliases should take precedence:\n%s", got)
	}
}

func TestBinaryRunDirectlyFallsBackQuietly(t *testing.T) {
	s := newSandbox(t)
	out := s.ali("add", "gp", "echo pulled") // no wrapper: no shell report, no failure
	if strings.Contains(out, "already an alias") {
		t.Errorf("unexpected shell warning without wrapper:\n%s", out)
	}
}

func TestRemovingAnOverrideRestoresOhMyZshAliasInTheSameShell(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte(omzRC), 0o644)
	s.ali("init", "zsh")

	got := s.zsh("ali add gp 'echo pulled'\nali update gp 'echo rebased'\ngp\nali rm gp\nalias gp")
	if !hasLine(got, "rebased") {
		t.Errorf("update not live:\n%s", got)
	}
	if !hasLine(got, "gp='git push'") {
		t.Errorf("Oh My Zsh's gp not restored after rm:\n%s", got)
	}
}
