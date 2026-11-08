package store

import (
	"os"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/model"
)

func seedExpansion(d *Data) {
	d.Set(alias.Alias{Name: "k", Command: "kubectl", Group: "kubectl"}, false)
	d.SetSub("kubectl", "g", "get")
	d.SetSub("kubectl", "n", "-n {{ns}}")
	d.SetShared("o", "-o wide")
	d.SetVar("ns", model.Variable{Values: map[string]string{"one": "namespace-1", "sys": "kube-system"}, Default: "one"})
	d.SetEnv("prod", model.Environment{Protected: true, Set: map[string]string{"ns": "prod"}})
}

func TestExpansionRoundTripAndVersion(t *testing.T) {
	s := newStore(t)
	if err := s.Update(func(d *Data) error { seedExpansion(d); return nil }); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(s.File())
	if !strings.HasPrefix(string(raw), "version = 2") {
		t.Errorf("want version 2 when expansion tables exist:\n%s", raw)
	}
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if d.Subs("kubectl")["n"] != "-n {{ns}}" || d.Shared()["o"] != "-o wide" {
		t.Errorf("subs lost: %v %v", d.Subs("kubectl"), d.Shared())
	}
	if v, ok := d.Var("ns"); !ok || v.Default != "one" || v.Values["sys"] != "kube-system" {
		t.Errorf("var lost: %+v", v)
	}
	if e, ok := d.Env("prod"); !ok || !e.Protected || e.Set["ns"] != "prod" {
		t.Errorf("env lost: %+v", e)
	}
}

func TestVersionOneWhenNoExpansionTables(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error { seedExpansion(d); return nil })
	s.Update(func(d *Data) error {
		d.DeleteSub("kubectl", "g")
		d.DeleteSub("kubectl", "n")
		d.DeleteShared("o")
		d.DeleteVar("ns")
		d.DeleteEnv("prod")
		return nil
	})
	raw, _ := os.ReadFile(s.File())
	if !strings.HasPrefix(string(raw), "version = 1") || strings.Contains(string(raw), "_subs") {
		t.Errorf("want a plain v1 store again:\n%s", raw)
	}
}

func TestInvalidExpansionTablesAreReported(t *testing.T) {
	cases := map[string]string{
		"bad tool":        "version = 2\n[_subs.Kubectl]\ng = \"get\"\n",
		"bad sub name":    "version = 2\n[_subs.kubectl]\n\"@g\" = \"get\"\n",
		"sub not string":  "version = 2\n[_subs.kubectl]\ng = 1\n",
		"empty sub":       "version = 2\n[_subs.kubectl]\ng = \"\"\n",
		"bad shared":      "version = 2\n[_shared]\n\"a b\" = \"x\"\n",
		"default not key": "version = 2\n[_vars.ns]\ndefault = \"two\"\n[_vars.ns.values]\none = \"x\"\n",
		"bad var name":    "version = 2\n[_vars.Ns.values]\none = \"x\"\n",
		"env unknown var": "version = 2\n[_envs.prod.set]\nns = \"x\"\n",
		"env bad name":    "version = 2\n[_vars.ns.values]\none = \"x\"\n[_envs.Prod.set]\nns = \"x\"\n",
		"env protected":   "version = 2\n[_vars.ns.values]\none = \"x\"\n[_envs.prod]\nprotected = \"yes\"\n",
		"version 3":       "version = 3\n",
	}
	for name, body := range cases {
		s := newStore(t)
		os.WriteFile(s.File(), []byte(body), 0o600)
		if _, err := s.Load(); err == nil {
			t.Errorf("%s: want error", name)
		} else if !strings.Contains(err.Error(), "aliases.toml") {
			t.Errorf("%s: error should name the file: %v", name, err)
		}
	}
}

func TestVersionOneStoreStillLoads(t *testing.T) {
	s := newStore(t)
	os.WriteFile(s.File(), []byte("version = 1\n[git]\ngp = \"git pull\"\n"), 0o600)
	if d, err := s.Load(); err != nil || len(d.All()) != 1 {
		t.Fatalf("%v", err)
	}
}

func TestUsesOfVariable(t *testing.T) {
	d := newData()
	seedExpansion(d)
	d.Set(alias.Alias{Name: "kgp", Command: "kubectl get pods -n {{ns}}", Group: "kubectl"}, false)
	uses := d.UsesOf("ns")
	want := []string{"alias kgp", "env prod", "sub kubectl n"}
	if strings.Join(uses, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v", uses, want)
	}
}
