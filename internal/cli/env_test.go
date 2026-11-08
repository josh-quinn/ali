package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/settings"
)

func seedEnvVars(e *env) {
	e.must("var", "add", "ns", "one=namespace-1", "sys=kube-system")
	e.must("var", "add", "ctx", "dev=dev-cluster")
}

func TestEnvLifecycle(t *testing.T) {
	e := newEnv(t)
	seedEnvVars(e)
	e.must("env", "add", "dev", "ns=one", "ctx=dev")
	e.must("env", "add", "prod", "ns=production", "--protect")
	e.must("env", "set", "dev", "ns=sys")
	e.must("env", "unset", "prod", "ns")
	e.must("env", "unprotect", "prod")
	e.must("env", "protect", "prod")
	d, _ := e.st.Load()
	dev, _ := d.Env("dev")
	prod, _ := d.Env("prod")
	if dev.Set["ns"] != "sys" || !prod.Protected || len(prod.Set) != 0 {
		t.Errorf("%+v %+v", dev, prod)
	}
	out := e.must("env", "show", "dev")
	if !strings.Contains(out, "ns   sys (kube-system)") || !strings.Contains(out, "ctx  dev (dev-cluster)") {
		t.Errorf("show:\n%s", out)
	}
	e.must("env", "rm", "prod")
	if _, ok := func() (any, bool) { d, _ := e.st.Load(); return d.Env("prod") }(); ok {
		t.Error("prod not removed")
	}
}

func TestEnvErrors(t *testing.T) {
	e := newEnv(t)
	seedEnvVars(e)
	e.must("env", "add", "dev", "ns=one")
	for _, args := range [][]string{
		{"env", "add", "dev"},            // exists
		{"env", "add", "Dev"},            // bad name
		{"env", "add", "x", "nope=1"},    // unknown variable
		{"env", "add", "x", "ns"},        // not VAR=VALUE
		{"env", "set", "nope", "ns=one"}, // unknown env
		{"env", "unset", "dev", "ctx"},   // not set
		{"env", "use", "nope"},           // unknown env
		{"env", "show", "nope"},
		{"env", "rm", "nope"},
	} {
		if _, _, err := e.run(args...); err == nil {
			t.Errorf("%q: want error", args)
		}
	}
}

func TestEnvUseGlobalAndCurrent(t *testing.T) {
	e := newEnv(t)
	seedEnvVars(e)
	e.must("env", "add", "dev", "ns=one")
	e.must("env", "add", "prod", "ns=production", "--protect")

	if out := e.must("env", "current"); !strings.Contains(out, "none") {
		t.Errorf("current: %q", out)
	}
	e.must("env", "use", "dev", "--global")
	if g, _ := settings.GlobalEnv(e.st.Dir); g != "dev" {
		t.Errorf("global = %q", g)
	}
	if out := e.must("env", "current"); !strings.Contains(out, "dev (global)") {
		t.Errorf("current: %q", out)
	}
	t.Setenv("ALI_ENV", "prod")
	if out := e.must("env", "current"); !strings.Contains(out, "prod (this terminal, protected)") {
		t.Errorf("current: %q", out)
	}
	ls := e.must("env", "ls")
	if !strings.Contains(ls, "* prod") || !strings.Contains(ls, "  dev") {
		t.Errorf("ls should mark the active one:\n%s", ls)
	}
	e.must("env", "off", "--global")
	if g, _ := settings.GlobalEnv(e.st.Dir); g != "" {
		t.Errorf("global not cleared: %q", g)
	}
}

func TestEnvUseWithoutWrapperExplains(t *testing.T) {
	e := newEnv(t)
	seedEnvVars(e)
	e.must("env", "add", "dev", "ns=one")
	t.Setenv("ALI_WRAPPER", "")
	_, errs, err := e.run("env", "use", "dev")
	if err != nil || !strings.Contains(errs, "--global") {
		t.Errorf("err=%v stderr=%q", err, errs)
	}
	t.Setenv("ALI_WRAPPER", "1")
	out, errs, _ := e.run("env", "use", "dev")
	if !strings.Contains(out, "using dev in this terminal") || errs != "" {
		t.Errorf("out=%q stderr=%q", out, errs)
	}
}

func TestEnvRmClearsGlobal(t *testing.T) {
	e := newEnv(t)
	seedEnvVars(e)
	e.must("env", "add", "dev", "ns=one")
	e.must("env", "use", "dev", "--global")
	_, errs, _ := e.run("env", "rm", "dev")
	if g, _ := settings.GlobalEnv(e.st.Dir); g != "" || !strings.Contains(errs, "global") {
		t.Errorf("global=%q stderr=%q", g, errs)
	}
	if b, _ := os.ReadFile(filepath.Join(e.st.Dir, settings.GlobalEnvFile)); strings.TrimSpace(string(b)) != "" {
		t.Errorf("global-env file: %q", b)
	}
}

func TestEnvJSON(t *testing.T) {
	e := newEnv(t)
	seedEnvVars(e)
	e.must("env", "add", "prod", "ns=production", "--protect")
	var rows []map[string]any
	json.Unmarshal([]byte(e.must("env", "ls", "--json")), &rows)
	if len(rows) != 1 || rows[0]["name"] != "prod" || rows[0]["protected"] != true {
		t.Errorf("%v", rows)
	}
}
