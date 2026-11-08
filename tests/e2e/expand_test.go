package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// expandSandbox has alias k=kubectl, kubectl sub-aliases g/p, a fake kubectl
// that prints its arguments, and an ali binary wrapped to log every call.
func expandSandbox(t *testing.T) (*sandbox, string) {
	s := newSandbox(t)
	s.ali("init", "zsh")
	s.ali("add", "k", "kubectl")
	s.ali("sub", "add", "kubectl", "g", "get")
	s.ali("sub", "add", "kubectl", "p", "pods")
	os.WriteFile(filepath.Join(s.bin, "kubectl"), []byte("#!/bin/sh\necho \"kubectl:$*\"\n"), 0o755)
	log := filepath.Join(s.home, "ali-calls.log")
	os.Rename(filepath.Join(s.bin, "ali"), filepath.Join(s.bin, "ali-real"))
	os.WriteFile(filepath.Join(s.bin, "ali"), []byte("#!/bin/sh\necho \"$*\" >> "+log+"\nexec "+filepath.Join(s.bin, "ali-real")+" \"$@\"\n"), 0o755)
	return s, log
}

func calls(t *testing.T, log string) string {
	b, _ := os.ReadFile(log)
	return string(b)
}

func TestTypedSubAliasesRunExpanded(t *testing.T) {
	s, _ := expandSandbox(t)
	got := s.zle("k g p\rk get p\rkubectl g p\r")
	if strings.Count(got, "kubectl:get pods") != 3 {
		t.Errorf("want 3 expanded runs:\n%s", got)
	}
}

func TestSpaceExpandsBeforeEnter(t *testing.T) {
	s, _ := expandSandbox(t)
	// type "k g p ", jump to the start (Ctrl-A) and prefix "echo ": if Space had
	// not expanded the line, echo would print "k g p".
	got := s.zle("k g p \x01echo \r")
	if !hasLine(got, "kubectl get pods") {
		t.Errorf("line was not expanded on Space:\n%s", got)
	}
}

func TestHistoryRecordsTheExpandedCommand(t *testing.T) {
	s, _ := expandSandbox(t)
	got := s.zle("k g p\rfc -ln -1 -1\r")
	if !hasLine(got, "kubectl get pods") {
		t.Errorf("history does not hold the expanded command:\n%s", got)
	}
}

func TestBackslashKeepsWordLiteral(t *testing.T) {
	s, _ := expandSandbox(t)
	got := s.zle("k g \\p\r")
	if !strings.Contains(got, "kubectl:get p") || strings.Contains(got, "kubectl:get pods") {
		t.Errorf(`\p should stay literal:\n%s`, got)
	}
}

func TestOrdinaryLinesNeverStartAProcess(t *testing.T) {
	s, log := expandSandbox(t)
	s.zle("ls -la >/dev/null\recho g p\rgit --version >/dev/null 2>&1\r")
	if strings.Contains(calls(t, log), "expand") {
		t.Errorf("ali was called for ordinary lines:\n%s", calls(t, log))
	}
	s.zle("k g p\r")
	if !strings.Contains(calls(t, log), "expand --zle") {
		t.Errorf("ali not called for an expandable line:\n%s", calls(t, log))
	}
}
