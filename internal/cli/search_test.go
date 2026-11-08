package cli

import (
	"strings"
	"testing"
)

func TestSearch(t *testing.T) {
	e := newEnv(t)
	seed(e)
	out := e.must("search", "PODS")
	if !strings.Contains(out, "hgp") || strings.Contains(out, "gp   git pull") {
		t.Errorf("search results:\n%s", out)
	}
	if out := e.must("search", "kube"); !strings.Contains(out, "kubectl") {
		t.Errorf("group/command match:\n%s", out)
	}
	if _, errs, err := e.run("search", "zzz"); err == nil || !strings.Contains(errs+err.Error(), "no aliases") {
		t.Errorf("no-match should fail: %v %q", err, errs)
	}
}
