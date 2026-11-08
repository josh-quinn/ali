package cli

import (
	"strings"
	"testing"
)

func seedSubs(e *env) {
	e.must("add", "k", "kubectl")
	e.must("sub", "add", "kubectl", "g", "get")
	e.must("sub", "add", "kubectl", "p", "pods")
}

func TestExpandPrintsWithoutRunning(t *testing.T) {
	e := newEnv(t)
	seedSubs(e)
	if out := e.must("expand", "k g p"); out != "kubectl get pods\n" {
		t.Errorf("got %q", out)
	}
	// nothing to expand: printed as typed, still success
	if out := e.must("expand", "ls -la"); out != "ls -la\n" {
		t.Errorf("got %q", out)
	}
	if _, _, err := e.run("expand"); err == nil {
		t.Error("missing line should be a usage error")
	}
}

func fields(out string) []string { return strings.Split(out, "\x00") }

func TestExpandZleProtocol(t *testing.T) {
	e := newEnv(t)
	seedSubs(e)
	f := fields(e.must("expand", "--zle", "space", "--", "k g "))
	if f[0] != "ok" || f[1] != "kubectl get " {
		t.Errorf("space: %q", f)
	}
	f = fields(e.must("expand", "--zle", "accept", "--", "k g p"))
	if f[0] != "ok" || f[1] != "kubectl get pods" {
		t.Errorf("accept: %q", f)
	}
	f = fields(e.must("expand", "--zle", "accept", "--", "ls"))
	if f[0] != "unchanged" {
		t.Errorf("unchanged: %q", f)
	}
	if _, _, err := e.run("expand", "--zle", "sideways", "--", "x"); err == nil {
		t.Error("bad --zle value accepted")
	}
}

func TestExpandZleNeverFailsTheWidgetOnStoreErrors(t *testing.T) {
	e := newEnv(t)
	e.must("add", "x", "y")
	writeFile(t, e.st.File(), "[broken\n")
	f := fields(e.must("expand", "--zle", "accept", "--", "k g p"))
	if f[0] != "error" || !strings.Contains(f[2], "aliases.toml") {
		t.Errorf("want error status naming the file, got %q", f)
	}
}
