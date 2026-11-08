package cli

import "testing"

func TestRegroup(t *testing.T) {
	e := newEnv(t)
	e.must("add", "gp", "git pull")
	e.must("regroup", "gp", "tools")
	d, _ := e.st.Load()
	if a, _ := d.Get("gp"); a.Group != "tools" {
		t.Errorf("%+v", a)
	}
	if _, _, err := e.run("regroup", "gp", "Bad Group"); err == nil {
		t.Error("bad group accepted")
	}
	if _, _, err := e.run("regroup", "nope", "x"); err == nil {
		t.Error("missing alias accepted")
	}
}
