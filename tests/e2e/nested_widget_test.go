package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Plugins such as fzf bind Tab (or other keys) to their own widget, which then
// calls "zle expand-or-complete" or "zle self-insert". ali's wrappers must hand
// off to the widget actually called, not the outer one bound to the key.
const outerWidgets = `_outer_tab() { zle expand-or-complete; }
zle -N _outer_tab
bindkey '^I' _outer_tab
_outer_space() { zle self-insert; }
zle -N _outer_space
bindkey ' ' _outer_space
`

func nestedSandbox(t *testing.T, rc string) *sandbox {
	t.Helper()
	s, _ := expandSandbox(t)
	b, _ := os.ReadFile(filepath.Join(s.home, ".zshrc"))
	os.WriteFile(filepath.Join(s.home, ".zshrc"), append([]byte(rc), b...), 0o644)
	os.MkdirAll(filepath.Join(s.home, "alpha-dir"), 0o755)
	return s
}

func checkNested(t *testing.T, s *sandbox) {
	t.Helper()
	got := s.zle("cd ~/alp\t\rpwd\rk g p \x01echo \r")
	if strings.Contains(got, "No such widget") {
		t.Fatalf("wrapper handed off to the wrong widget:\n%s", got)
	}
	if !strings.Contains(got, "alpha-dir") {
		t.Errorf("ordinary Tab completion through the outer widget failed:\n%s", got)
	}
	if !hasLine(got, "kubectl get pods") {
		t.Errorf("Space expansion through the outer widget failed:\n%s", got)
	}
}

func TestWrappersWorkWhenCalledFromAnotherWidget(t *testing.T) {
	checkNested(t, nestedSandbox(t, outerWidgets))
}

func TestWorksWithFzf(t *testing.T) {
	fzf, err := exec.LookPath("fzf")
	if err != nil {
		t.Skip("fzf not installed")
	}
	s := nestedSandbox(t, "source <(fzf --zsh)\n")
	os.Symlink(fzf, filepath.Join(s.bin, "fzf"))
	if got := s.zle("bindkey '^I'\r"); !strings.Contains(got, "fzf-completion") {
		t.Fatalf("fzf did not load in the sandbox:\n%s", got)
	}
	checkNested(t, s)
}
