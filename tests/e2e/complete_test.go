package e2e

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestTabCompletesAUniqueValueAndRunsIt(t *testing.T) {
	s := placeholderSandbox(t)
	got := s.zle("k g p n @s\t\r")
	if !strings.Contains(got, "[get][pods][-n][kube-system]") {
		t.Errorf("@s<Tab> should complete to @sys:\n%s", got)
	}
}

func TestTabListsValuesAndSubAliases(t *testing.T) {
	s := placeholderSandbox(t)
	got := s.zle("k g p n @\t" + waitFor("@sys") + "\x15k \t" + waitFor("n  p") + "\x15echo done\r")
	if !regexp.MustCompile(`@one\s+@sys`).MatchString(got) {
		t.Errorf("values not listed:\n%s", got)
	}
	if !regexp.MustCompile(`g\s+n\s+p`).MatchString(got) {
		t.Errorf("kubectl sub-aliases not listed:\n%s", got)
	}
	if !hasLine(got, "done") {
		t.Errorf("shell not usable after completion:\n%s", got)
	}
}

func TestTabCompletesAlisOwnCommands(t *testing.T) {
	s := newSandbox(t)
	os.WriteFile(filepath.Join(s.home, ".zshrc"), []byte("autoload -Uz compinit && compinit -u\n"), 0o644)
	s.ali("init", "zsh")
	got := s.zle("ali sh\t--help\r")
	if !strings.Contains(got, "Show, for TOOL") {
		t.Errorf("'ali sh<Tab>' did not complete to 'ali show':\n%s", got)
	}
}
