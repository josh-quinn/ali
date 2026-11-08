package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigGetSetAndRegenerate(t *testing.T) {
	e := newEnv(t)
	e.must("add", "k", "kubectl")
	e.must("sub", "add", "kubectl", "g", "get")
	if out := e.must("config", "get", "expand"); strings.TrimSpace(out) != "inline" {
		t.Errorf("default expand: %q", out)
	}
	z, _ := os.ReadFile(e.st.ZshFile())
	if !strings.Contains(string(z), "_ali_accept") {
		t.Fatal("inline mode should generate widgets")
	}
	e.must("config", "set", "expand", "runtime")
	if out := e.must("config", "get", "expand"); strings.TrimSpace(out) != "runtime" {
		t.Errorf("after set: %q", out)
	}
	z, _ = os.ReadFile(e.st.ZshFile())
	if strings.Contains(string(z), "_ali_accept") || !strings.Contains(string(z), "functions[k]=") {
		t.Errorf("runtime mode not regenerated:\n%s", z)
	}
	if out := e.must("config", "ls"); !strings.Contains(out, "expand = runtime") || !strings.Contains(out, "prompt = right") {
		t.Errorf("ls: %q", out)
	}
}

func TestConfigErrors(t *testing.T) {
	e := newEnv(t)
	for _, args := range [][]string{{"config", "set", "expand", "sometimes"}, {"config", "set", "colour", "red"}, {"config", "get", "nope"}} {
		if _, _, err := e.run(args...); err == nil {
			t.Errorf("%q: want error", args)
		}
	}
	os.MkdirAll(e.st.Dir, 0o700)
	os.WriteFile(filepath.Join(e.st.Dir, "settings.toml"), []byte("expand = 1\n"), 0o600)
	if _, _, err := e.run("config", "ls"); err == nil || !strings.Contains(err.Error(), "settings.toml") {
		t.Errorf("broken settings should be reported: %v", err)
	}
}
