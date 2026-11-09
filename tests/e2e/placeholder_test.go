package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// placeholderSandbox extends expandSandbox with ns and a fake kubectl that
// prints each argument on its own line between brackets.
func placeholderSandbox(t *testing.T) *sandbox {
	s, _ := expandSandbox(t)
	s.ali("var", "add", "ns", "one=namespace-1", "sys=kube-system", "--default", "one")
	s.ali("sub", "add", "kubectl", "n", "-n {{ns}}")
	s.ali("add", "kgp", "kubectl get pods -n {{ns}}")
	os.WriteFile(filepath.Join(s.bin, "kubectl"), []byte("#!/bin/sh\nfor a in \"$@\"; do printf '[%s]' \"$a\"; done; echo\n"), 0o755)
	return s
}

func TestPlaceholdersInTheLineEditor(t *testing.T) {
	s := placeholderSandbox(t)
	got := s.zle("k g p n @sys\rk g p n\rkgp @ns=x-y\r")
	for _, want := range []string{
		"[get][pods][-n][kube-system]",
		"[get][pods][-n][namespace-1]",
		"[get][pods][-n][x-y]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s:\n%s", want, got)
		}
	}
}

func TestHostileValueArrivesAsOneArgument(t *testing.T) {
	s := placeholderSandbox(t)
	got := s.zle("k g p n @ns='a b; touch PWNED'\r")
	if !strings.Contains(got, "[get][pods][-n][a b; touch PWNED]") {
		t.Errorf("value not passed as one literal argument:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(s.home, "PWNED")); err == nil {
		t.Error("value was executed")
	}
}

func TestUnresolvedPlaceholderDoesNotRun(t *testing.T) {
	s := placeholderSandbox(t)
	s.ali("var", "add", "ctx", "dev=dev-cluster")
	s.ali("sub", "add", "kubectl", "c", "--context={{ctx}}")
	// wait for ali's message before Ctrl-U clears the line it kept in the editor
	got := s.zle("k g p c\r" + waitFor("no value for") + "\x15echo after\r")
	if strings.Contains(got, "[get]") {
		t.Errorf("line ran without a value for ctx:\n%s", got)
	}
	if !strings.Contains(got, "no value for {{ctx}}") || !hasLine(got, "after") {
		t.Errorf("want message and a usable editor:\n%s", got)
	}
}

// Without the line editor (scripts, piped input) an alias with a placeholder
// runs as a function and still expands.
func TestPlaceholderAliasWithoutTheLineEditor(t *testing.T) {
	s := placeholderSandbox(t)
	got := s.zsh("kgp @sys")
	if !strings.Contains(got, "[get][pods][-n][kube-system]") {
		t.Errorf("function fallback failed:\n%s", got)
	}
}
