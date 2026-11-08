package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShowTool(t *testing.T) {
	e := newEnv(t)
	e.must("var", "add", "ns", "one=namespace-1", "--default", "one")
	e.must("add", "k", "kubectl")
	e.must("add", "kgp", "kubectl get pods -n {{ns}}")
	e.must("add", "gp", "git pull")
	e.must("sub", "add", "kubectl", "g", "get")
	e.must("sub", "add", "kubectl", "o", "-o yaml")
	e.must("sub", "add", "--shared", "o", "-o wide")
	e.must("sub", "add", "--shared", "w", "--watch")
	out := e.must("show", "kubectl")
	for _, want := range []string{"aliases", "k    kubectl", "kgp  kubectl get pods -n {{ns}}", "sub-aliases", "g  get", "o  -o yaml",
		"shared sub-aliases", "w  --watch", "o  -o wide  (overridden by kubectl's o)", "variables", "ns"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "git pull") {
		t.Errorf("other tools' aliases listed:\n%s", out)
	}
	var j map[string]any
	if err := json.Unmarshal([]byte(e.must("show", "kubectl", "--json")), &j); err != nil || j["tool"] != "kubectl" {
		t.Errorf("%v %v", err, j)
	}
	if _, _, err := e.run("show", "nothing-here"); err == nil {
		t.Error("unknown tool should be an error")
	}
}
