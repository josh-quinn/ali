package cli

import (
	"errors"
	"io"
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

func TestExpandRunQuotesWordsAndAlwaysExpandsTheAlias(t *testing.T) {
	e := newEnv(t)
	seedSubs(e)
	e.must("var", "add", "ns", "sys=kube-system", "--default", "sys")
	e.must("add", "kgp", "kubectl get pods -n {{ns}}")
	cases := map[string][]string{
		"kubectl get pods -n kube-system":         {"kgp"},
		"kubectl get pods -n 'a b'":               {"kgp", "@ns=a b"},
		"kubectl logs 'my pod'":                   {"k", "logs", "my pod"}, // nothing to expand: k still expands, no recursion
		`kubectl get \p`:                          {"k", "g", `\p`},
		"kubectl get pods -n kube-system '$(id)'": {"kgp", "$(id)"},
	}
	for want, words := range cases {
		out := e.must(append([]string{"expand", "--run", "--"}, words...)...)
		if strings.TrimSpace(out) != want {
			t.Errorf("%q => %q, want %q", words, strings.TrimSpace(out), want)
		}
	}
}

type fakeTTY struct {
	in  *strings.Reader
	out strings.Builder
}

func (f *fakeTTY) Read(p []byte) (int, error)  { return f.in.Read(p) }
func (f *fakeTTY) Write(p []byte) (int, error) { return f.out.Write(p) }
func (f *fakeTTY) Close() error                { return nil }

func withTTY(t *testing.T, answer string, fail bool) *fakeTTY {
	t.Helper()
	f := &fakeTTY{in: strings.NewReader(answer)}
	old := openTTY
	openTTY = func() (io.ReadWriteCloser, error) {
		if fail {
			return nil, errors.New("no tty")
		}
		return f, nil
	}
	t.Cleanup(func() { openTTY = old })
	return f
}

func TestRunConfirmsProtectedEnvironmentsOnTheTerminal(t *testing.T) {
	e := newEnv(t)
	e.must("var", "add", "ns", "sys=kube-system")
	e.must("env", "add", "prod", "ns=production", "--protect")
	e.must("add", "kgp", "kubectl get pods -n {{ns}}")
	t.Setenv("ALI_ENV", "prod")

	tty := withTTY(t, "y\n", false)
	out := e.must("expand", "--run", "--", "kgp")
	if strings.TrimSpace(out) != "kubectl get pods -n production" || !strings.Contains(tty.out.String(), "protected") {
		t.Errorf("out=%q tty=%q", out, tty.out.String())
	}
	withTTY(t, "n\n", false)
	if out, _, err := e.run("expand", "--run", "--", "kgp"); err == nil || out != "" {
		t.Errorf("declined command must not be printed for eval: out=%q err=%v", out, err)
	}
	withTTY(t, "", true)
	if _, _, err := e.run("expand", "--run", "--", "kgp"); err == nil || !strings.Contains(err.Error(), "no terminal") {
		t.Errorf("without a terminal it must refuse: %v", err)
	}
	// an explicit value does not need confirmation
	withTTY(t, "", true)
	if out := e.must("expand", "--run", "--", "kgp", "@sys"); strings.TrimSpace(out) != "kubectl get pods -n kube-system" {
		t.Errorf("%q", out)
	}
}
