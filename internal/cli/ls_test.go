package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLsGrouped(t *testing.T) {
	e := newEnv(t)
	seed(e)
	out := e.must("ls")
	gi, hi := strings.Index(out, "git\n"), strings.Index(out, "helm\n")
	if gi < 0 || hi < 0 || gi > hi {
		t.Errorf("groups missing/unsorted:\n%s", out)
	}
	if !strings.Contains(out, "  gp   git pull") || !strings.Contains(out, "  hgp  helm get pods") {
		t.Errorf("not aligned:\n%s", out)
	}
	only := e.must("ls", "helm")
	if strings.Contains(only, "git pull") || !strings.Contains(only, "helm get pods") {
		t.Errorf("group filter wrong:\n%s", only)
	}
	if _, _, err := e.run("ls", "nope"); err == nil {
		t.Error("unknown group should error")
	}
}

func TestLsEmptyAndJSON(t *testing.T) {
	e := newEnv(t)
	if out := e.must("ls"); !strings.Contains(out, "ali add") {
		t.Errorf("empty hint missing: %q", out)
	}
	if out := e.must("ls", "--json"); strings.TrimSpace(out) != "[]" {
		t.Errorf("empty json: %q", out)
	}
	seed(e)
	var got []map[string]string
	if err := json.Unmarshal([]byte(e.must("ls", "--json")), &got); err != nil || len(got) != 4 {
		t.Fatalf("json: %v %v", err, got)
	}
	if got[0]["name"] == "" || got[0]["command"] == "" || got[0]["group"] == "" {
		t.Errorf("json fields: %v", got[0])
	}
}

// A stored command (or a hand-edited store) must not be able to drive the
// terminal: escape sequences could clear the screen or hide text.
func TestListingNeutralisesTerminalControlSequences(t *testing.T) {
	e := newEnv(t)
	e.must("add", "evil", "echo \x1b[2J\x1b]0;pwned\a\x1b[8mhidden\r\u009b31m")
	for _, args := range [][]string{{"ls"}, {"search", "evil"}, {"add", "evil2", "echo \x1b[2J"}, {"update", "evil", "echo \x1b[31m"}} {
		out := e.must(args...)
		for _, r := range out {
			if r == '\x1b' || r == '\a' || r == '\r' || (r >= 0x80 && r <= 0x9f) {
				t.Fatalf("ali %v leaked control character %q in %q", args, r, out)
			}
		}
	}
	if out := e.must("ls"); !strings.Contains(out, `\x1b[2J`) {
		t.Errorf("escape should be shown visibly:\n%s", out)
	}
	// the stored value itself must stay verbatim
	d, _ := e.st.Load()
	if a, _ := d.Get("evil2"); a.Command != "echo \x1b[2J" {
		t.Errorf("stored command was altered: %q", a.Command)
	}
}
