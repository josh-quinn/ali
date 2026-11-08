package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGroups(t *testing.T) {
	e := newEnv(t)
	seed(e)
	out := e.must("groups")
	for _, want := range []string{"git", "helm", "kubectl"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "git      2") {
		t.Errorf("count/alignment wrong:\n%s", out)
	}
	var got []map[string]any
	json.Unmarshal([]byte(e.must("groups", "--json")), &got)
	if len(got) != 3 {
		t.Errorf("json groups: %v", got)
	}
}
