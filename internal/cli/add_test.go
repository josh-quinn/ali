package cli

import (
	"os"
	"strings"
	"testing"
)

func TestAddInfersGroupAndPersists(t *testing.T) {
	e := newEnv(t)
	out := e.must("add", "gp", "git pull")
	if !strings.Contains(out, "added gp (git): git pull") {
		t.Errorf("output %q", out)
	}
	d, _ := e.st.Load()
	if a, ok := d.Get("gp"); !ok || a.Group != "git" || a.Command != "git pull" {
		t.Errorf("stored %+v", a)
	}
	if z, _ := os.ReadFile(e.st.ZshFile()); !strings.Contains(string(z), "alias -- gp='git pull'") {
		t.Errorf("zsh file: %s", z)
	}
}

func TestAddExplicitGroup(t *testing.T) {
	e := newEnv(t)
	e.must("add", "hgp", "helm get pods", "-g", "k8s")
	d, _ := e.st.Load()
	if a, _ := d.Get("hgp"); a.Group != "k8s" || !d.Pinned("hgp") {
		t.Errorf("%+v pinned=%v", a, d.Pinned("hgp"))
	}
}

func TestAddExistsNeedsForce(t *testing.T) {
	e := newEnv(t)
	e.must("add", "gp", "git pull")
	_, errs, err := e.run("add", "gp", "git push")
	if err == nil || !strings.Contains(err.Error(), "already exists") || !strings.Contains(err.Error(), "ali update") {
		t.Errorf("want exists error, got %v / %q", err, errs)
	}
	d, _ := e.st.Load()
	if a, _ := d.Get("gp"); a.Command != "git pull" {
		t.Error("alias changed without --force")
	}
	e.must("add", "gp", "git push", "--force")
	d, _ = e.st.Load()
	if a, _ := d.Get("gp"); a.Command != "git push" {
		t.Error("--force did not overwrite")
	}
}

func TestAddValidation(t *testing.T) {
	e := newEnv(t)
	for _, args := range [][]string{
		{"add", "bad name", "x"},
		{"add", "-x", "x"},
		{"add", "ok", "  "},
		{"add", "ok", "x", "-g", "Bad Group"},
	} {
		if _, _, err := e.run(args...); err == nil {
			t.Errorf("%v: want error", args)
		}
	}
	if d, _ := e.st.Load(); len(d.All()) != 0 {
		t.Error("invalid input was stored")
	}
}

func TestAddWarnsOnShadow(t *testing.T) {
	e := newEnv(t)
	_, errs, err := e.run("add", "cd", "cd ~/work")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(errs, "warning") || !strings.Contains(errs, "cd") {
		t.Errorf("no shadow warning: %q", errs)
	}
	d, _ := e.st.Load()
	if _, ok := d.Get("cd"); !ok {
		t.Error("shadowing alias should still be created")
	}
}

func TestAddSpecialCharsStoredVerbatim(t *testing.T) {
	e := newEnv(t)
	cmd := "echo 'it'\\''s' \"$HOME\" `x` && y | z\nnext é"
	e.must("add", "sp", cmd)
	d, _ := e.st.Load()
	if a, _ := d.Get("sp"); a.Command != cmd {
		t.Errorf("got %q", a.Command)
	}
}

func TestUsageErrors(t *testing.T) {
	e := newEnv(t)
	_, _, err := e.run("add", "onlyname")
	var u *UsageError
	if err == nil || !asUsage(err, &u) {
		t.Errorf("want usage error, got %T %v", err, err)
	}
}

func TestBrokenStoreIsReportedNotOverwritten(t *testing.T) {
	e := newEnv(t)
	os.MkdirAll(e.st.Dir, 0o755)
	os.WriteFile(e.st.File(), []byte("[oops\n"), 0o644)
	if _, _, err := e.run("add", "a", "b"); err == nil || !strings.Contains(err.Error(), "aliases.toml") {
		t.Errorf("want error naming file: %v", err)
	}
	if b, _ := os.ReadFile(e.st.File()); string(b) != "[oops\n" {
		t.Error("broken store overwritten")
	}
	if _, _, err := e.run("ls"); err == nil {
		t.Error("ls on broken store should fail")
	}
}
