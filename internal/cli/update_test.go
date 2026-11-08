package cli

import "testing"

func TestUpdateKeepsOrReinfersGroup(t *testing.T) {
	e := newEnv(t)
	e.must("add", "x", "git pull")
	e.must("update", "x", "svn up")
	d, _ := e.st.Load()
	if a, _ := d.Get("x"); a.Group != "svn" || a.Command != "svn up" {
		t.Errorf("not re-inferred: %+v", a)
	}
	e.must("regroup", "x", "vcs")
	e.must("update", "x", "hg pull")
	d, _ = e.st.Load()
	if a, _ := d.Get("x"); a.Group != "vcs" {
		t.Errorf("pinned group lost: %+v", a)
	}
	if _, _, err := e.run("update", "nope", "x"); err == nil {
		t.Error("missing alias should error")
	}
}
