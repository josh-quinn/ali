package cli

import (
	"strings"
	"testing"
)

func TestRmMultipleWithPartialFailure(t *testing.T) {
	e := newEnv(t)
	seed(e)
	out, _, err := e.run("rm", "gp", "nope", "gs")
	if err == nil || !strings.Contains(err.Error(), "nope") {
		t.Errorf("want not-found error naming nope, got %v", err)
	}
	if !strings.Contains(out, "removed gp: git pull") {
		t.Errorf("output %q", out)
	}
	d, _ := e.st.Load()
	if _, ok := d.Get("gp"); ok {
		t.Error("gp not removed")
	}
	if _, ok := d.Get("gs"); ok {
		t.Error("gs not removed despite partial failure")
	}
	if _, ok := d.Get("hgp"); !ok {
		t.Error("unrelated alias removed")
	}
}
