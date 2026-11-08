package cli

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	e := newEnv(t)
	if out := e.must("--version"); !strings.Contains(out, "test") {
		t.Errorf("version: %q", out)
	}
}
