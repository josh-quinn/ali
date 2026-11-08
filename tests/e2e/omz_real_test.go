package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realOMZ sets up the developer's own Oh My Zsh (if installed) in the sandbox,
// with updates disabled and every cache file redirected into the sandbox, so
// nothing under ~/.oh-my-zsh is written. Skipped when Oh My Zsh is absent.
func realOMZ(t *testing.T) *sandbox {
	t.Helper()
	home, _ := os.UserHomeDir()
	omz := filepath.Join(home, ".oh-my-zsh")
	if _, err := os.Stat(filepath.Join(omz, "oh-my-zsh.sh")); err != nil {
		t.Skip("Oh My Zsh not installed")
	}
	s, _ := expandSandbox(t)
	cache := filepath.Join(s.home, "omz-cache")
	os.MkdirAll(cache, 0o700)
	rc := "export ZSH=" + omz + "\nexport ZSH_CACHE_DIR=" + cache + "\nexport ZSH_COMPDUMP=" + cache + "/zcompdump\n" +
		"DISABLE_AUTO_UPDATE=true\nzstyle ':omz:update' mode disabled\nplugins=(git)\nsource $ZSH/oh-my-zsh.sh\n"
	b, _ := os.ReadFile(filepath.Join(s.home, ".zshrc"))
	os.WriteFile(filepath.Join(s.home, ".zshrc"), append([]byte(rc), b...), 0o644)
	return s
}

func TestWithRealOhMyZsh(t *testing.T) {
	s := realOMZ(t)
	got := s.zle("k g p \x01echo \rk g p\rbindkey ' '\rls /us\t\recho done\r") // Enter after Tab keeps the listing on screen
	if !hasLine(got, "kubectl get pods") {
		t.Errorf("Space expansion broken under Oh My Zsh:\n%s", got)
	}
	if !strings.Contains(got, "kubectl:get pods") {
		t.Errorf("Enter expansion broken under Oh My Zsh:\n%s", got)
	}
	if !strings.Contains(got, "magic-space") {
		t.Errorf("Oh My Zsh's Space binding should be left in place:\n%s", got)
	}
	if !strings.Contains(got, "usr/") { // Oh My Zsh lists Users/ and usr/ (case-insensitive)
		t.Errorf("ordinary Tab completion broken under Oh My Zsh:\n%s", got)
	}
	if !hasLine(got, "done") {
		t.Errorf("shell not usable:\n%s", got)
	}
}
