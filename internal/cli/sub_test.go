package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSubAddListUpdateRemove(t *testing.T) {
	e := newEnv(t)
	if out := e.must("sub", "add", "kubectl", "g", "get"); !strings.Contains(out, "added sub-alias g for kubectl: get") {
		t.Errorf("add output %q", out)
	}
	e.must("sub", "add", "kubectl", "p", "pods")
	e.must("sub", "add", "helm", "g", "get values")
	if out := e.must("sub", "add", "--shared", "o", "-o wide"); !strings.Contains(out, "added shared sub-alias o: -o wide") {
		t.Errorf("shared add output %q", out)
	}

	out := e.must("sub", "ls")
	for _, want := range []string{"helm\n  g  get values", "kubectl\n  g  get\n  p  pods", "shared\n  o  -o wide"} {
		if !strings.Contains(out, want) {
			t.Errorf("ls missing %q:\n%s", want, out)
		}
	}
	if only := e.must("sub", "ls", "helm"); strings.Contains(only, "pods") || !strings.Contains(only, "get values") {
		t.Errorf("ls helm:\n%s", only)
	}
	if only := e.must("sub", "ls", "--shared"); !strings.Contains(only, "-o wide") || strings.Contains(only, "pods") {
		t.Errorf("ls --shared:\n%s", only)
	}

	e.must("sub", "update", "kubectl", "p", "pod")
	e.must("sub", "rm", "helm", "g")
	e.must("sub", "rm", "--shared", "o")
	d, _ := e.st.Load()
	if d.Subs("kubectl")["p"] != "pod" || len(d.Subs("helm")) != 0 || len(d.Shared()) != 0 {
		t.Errorf("store: %v %v %v", d.Subs("kubectl"), d.Subs("helm"), d.Shared())
	}
}

func TestSubErrors(t *testing.T) {
	e := newEnv(t)
	e.must("sub", "add", "kubectl", "g", "get")
	cases := [][]string{
		{"sub", "add", "kubectl", "g", "got"},     // exists without --force
		{"sub", "add", "Kubectl", "g", "get"},     // bad tool
		{"sub", "add", "kubectl", "@g", "get"},    // bad name
		{"sub", "add", "kubectl", "x", " "},       // empty expansion
		{"sub", "add", "kubectl", "x", "a\x00b"},  // NUL
		{"sub", "update", "kubectl", "nope", "x"}, // missing
		{"sub", "rm", "kubectl", "nope"},          // missing
		{"sub", "add", "--shared", "o"},           // wrong arg count
		{"sub", "add", "kubectl", "g"},            // wrong arg count
		{"sub", "ls", "nope"},                     // unknown tool
	}
	for _, args := range cases {
		if _, _, err := e.run(args...); err == nil {
			t.Errorf("%q: want error", args)
		}
	}
	e.must("sub", "add", "kubectl", "g", "got", "--force")
	if d, _ := e.st.Load(); d.Subs("kubectl")["g"] != "got" {
		t.Error("--force did not overwrite")
	}
}

func TestSubLsJSON(t *testing.T) {
	e := newEnv(t)
	if out := e.must("sub", "ls", "--json"); strings.TrimSpace(out) != "[]" {
		t.Errorf("empty: %q", out)
	}
	e.must("sub", "add", "kubectl", "g", "get")
	e.must("sub", "add", "--shared", "o", "-o wide")
	var rows []map[string]any
	if err := json.Unmarshal([]byte(e.must("sub", "ls", "--json")), &rows); err != nil || len(rows) != 2 {
		t.Fatalf("%v %v", err, rows)
	}
	if rows[0]["tool"] != "kubectl" || rows[1]["shared"] != true {
		t.Errorf("%v", rows)
	}
}

func TestSubAppearsInGeneratedFileRefresh(t *testing.T) {
	e := newEnv(t)
	e.must("sub", "add", "kubectl", "g", "get")
	if out := e.must("sub", "ls"); strings.ContainsRune(out, '\x1b') {
		t.Error("unsanitised output")
	}
}

// Expansions often start with "-" ("-o wide", "-n {{ns}}"); they must not be
// mistaken for flags, wherever the real flags appear.
func TestSubExpansionsMayStartWithADash(t *testing.T) {
	e := newEnv(t)
	e.must("sub", "add", "--shared", "o", "-o wide")
	e.must("sub", "add", "kubectl", "n", "-n {{ns}}", "--force")
	e.must("sub", "add", "kubectl", "a", "-A")
	e.must("sub", "update", "kubectl", "a", "--all-namespaces")
	d, _ := e.st.Load()
	if d.Shared()["o"] != "-o wide" || d.Subs("kubectl")["n"] != "-n {{ns}}" || d.Subs("kubectl")["a"] != "--all-namespaces" {
		t.Errorf("%v %v", d.Shared(), d.Subs("kubectl"))
	}
	if out := e.must("sub", "add", "--help"); !strings.Contains(out, "Add sub-alias NAME") {
		t.Errorf("help not shown: %q", out)
	}
}
