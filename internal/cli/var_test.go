package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVarLifecycle(t *testing.T) {
	e := newEnv(t)
	if out := e.must("var", "add", "ns", "one=namespace-1", "sys=kube-system", "--default", "one"); !strings.Contains(out, "added variable ns") {
		t.Errorf("%q", out)
	}
	e.must("var", "set", "ns", "dev=dev-ns", "sys=kube-system-2")
	e.must("var", "unset", "ns", "dev")
	e.must("var", "default", "ns", "sys")
	d, _ := e.st.Load()
	v, _ := d.Var("ns")
	if v.Default != "sys" || v.Values["sys"] != "kube-system-2" || len(v.Values) != 2 {
		t.Errorf("%+v", v)
	}
	e.must("var", "default", "ns", "--none")
	if d, _ := e.st.Load(); func() string { v, _ := d.Var("ns"); return v.Default }() != "" {
		t.Error("--none did not clear the default")
	}

	ls := e.must("var", "ls")
	if !strings.Contains(ls, "ns") || !strings.Contains(ls, "one  namespace-1") {
		t.Errorf("ls:\n%s", ls)
	}
}

func TestVarShowListsUsesAndRmIsRefusedWhileUsed(t *testing.T) {
	e := newEnv(t)
	e.must("var", "add", "ns", "one=namespace-1", "--default", "one")
	e.must("add", "kgp", "kubectl get pods -n {{ns}}")
	e.must("sub", "add", "kubectl", "n", "-n {{ns}}")
	out := e.must("var", "show", "ns")
	for _, want := range []string{"one  namespace-1  (default)", "alias kgp", "sub kubectl n"} {
		if !strings.Contains(out, want) {
			t.Errorf("show missing %q:\n%s", want, out)
		}
	}
	_, _, err := e.run("var", "rm", "ns")
	if err == nil || !strings.Contains(err.Error(), "alias kgp") {
		t.Errorf("rm of a used variable should be refused naming its uses: %v", err)
	}
	e.must("rm", "kgp")
	e.must("sub", "rm", "kubectl", "n")
	e.must("var", "rm", "ns")
}

func TestVarErrors(t *testing.T) {
	e := newEnv(t)
	e.must("var", "add", "ns", "one=x")
	for _, args := range [][]string{
		{"var", "add", "ns", "two=y"},                  // exists
		{"var", "add", "Ns", "one=x"},                  // bad name
		{"var", "add", "env", "one=x"},                 // reserved
		{"var", "add", "ctx", "Bad=x"},                 // bad key
		{"var", "add", "ctx", "novalue"},               // not KEY=VALUE
		{"var", "add", "ctx", "k="},                    // empty value
		{"var", "add", "ctx", "a=1", "--default", "b"}, // default not a key
		{"var", "set", "nope", "a=1"},                  // missing var
		{"var", "unset", "ns", "zzz"},                  // missing key
		{"var", "default", "ns", "zzz"},                // missing key
		{"var", "show", "nope"},
		{"var", "rm", "nope"},
	} {
		if _, _, err := e.run(args...); err == nil {
			t.Errorf("%q: want error", args)
		}
	}
}

func TestVarJSON(t *testing.T) {
	e := newEnv(t)
	if out := e.must("var", "ls", "--json"); strings.TrimSpace(out) != "[]" {
		t.Errorf("%q", out)
	}
	e.must("var", "add", "ns", "one=namespace-1", "--default", "one")
	var rows []map[string]any
	json.Unmarshal([]byte(e.must("var", "ls", "--json")), &rows)
	if len(rows) != 1 || rows[0]["name"] != "ns" || rows[0]["default"] != "one" {
		t.Errorf("%v", rows)
	}
}
