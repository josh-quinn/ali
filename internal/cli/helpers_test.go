package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/store"
)

type env struct {
	t   *testing.T
	st  *store.Store
	dir string
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("ZDOTDIR", "")
	t.Setenv("PATH", "/nonexistent")
	return &env{t: t, st: store.New(filepath.Join(dir, "cfg")), dir: dir}
}

// run executes ali with args and returns stdout, stderr and the error.
func (e *env) run(args ...string) (string, string, error) {
	return e.runIn("", args...)
}

// runIn is run with stdin.
func (e *env) runIn(stdin string, args ...string) (string, string, error) {
	var out, errb bytes.Buffer
	root := NewRoot("test", e.st, &out, &errb)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), errb.String(), err
}

func (e *env) must(args ...string) string {
	e.t.Helper()
	out, errs, err := e.run(args...)
	if err != nil {
		e.t.Fatalf("ali %v: %v (stderr %q)", args, err, errs)
	}
	return out
}

func seed(e *env) {
	e.must("add", "gp", "git pull")
	e.must("add", "gs", "git status")
	e.must("add", "hgp", "helm get pods")
	e.must("add", "k", "kubectl")
}
