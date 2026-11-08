package doctor

import (
	"errors"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/render"
)

func byName(cs []Check) map[string]Check {
	m := map[string]Check{}
	for _, c := range cs {
		m[c.Name] = c
	}
	return m
}

const omz = "export ZSH=$HOME/.oh-my-zsh\nplugins=(git)\nsource $ZSH/oh-my-zsh.sh\n"

func TestHealthySetup(t *testing.T) {
	cs := Run(Input{RC: omz + render.Block(), Overridden: map[string]string{}})
	for _, c := range cs {
		if c.Status != OK {
			t.Errorf("%s: %v %s", c.Name, c.Status, c.Detail)
		}
	}
	if Failed(cs) {
		t.Error("Failed() true for healthy setup")
	}
}

func TestEachProblemIsFlagged(t *testing.T) {
	cases := []struct {
		name  string
		in    Input
		check string
		want  Status
	}{
		{"missing block", Input{RC: omz}, CheckBlockPresent, Fail},
		{"damaged block", Input{RC: render.BlockStart + "\n"}, CheckBlockPresent, Fail},
		{"outdated block", Input{RC: render.BlockStart + "\nold\n" + render.BlockEnd + "\n"}, CheckBlockCurrent, Fail},
		{"block before omz", Input{RC: render.Block() + omz}, CheckOrder, Fail},
		{"no omz", Input{RC: render.Block()}, CheckOrder, Skipped},
		{"omz only in a comment", Input{RC: render.Block() + "# source $ZSH/oh-my-zsh.sh\n"}, CheckOrder, Skipped},
		{"bad store", Input{RC: render.Block(), StoreErr: errors.New("aliases.toml: bad")}, CheckStore, Fail},
		{"redefined", Input{RC: render.Block(), Overridden: map[string]string{"gp": "git push"}}, CheckRedefined, Fail},
		{"outside wrapper", Input{RC: render.Block()}, CheckRedefined, Skipped},
		{"no rc file", Input{RCMissing: true}, CheckBlockPresent, Fail},
	}
	for _, c := range cases {
		got := byName(Run(c.in))[c.check]
		if got.Status != c.want {
			t.Errorf("%s: %s = %v (%s), want %v", c.name, c.check, got.Status, got.Detail, c.want)
		}
		if got.Status == Fail && got.Fix == "" {
			t.Errorf("%s: failure without a fix hint", c.name)
		}
	}
}

func TestRedefinedListsNamesAndLiveValues(t *testing.T) {
	c := byName(Run(Input{RC: render.Block(), Overridden: map[string]string{"gp": "git push", "x": ""}}))[CheckRedefined]
	if !strings.Contains(c.Detail, "gp (now: git push)") || !strings.Contains(c.Detail, "x (now: not an alias)") {
		t.Errorf("detail: %q", c.Detail)
	}
}

func TestParseState(t *testing.T) {
	gen, m := ParseState("abc\x00gp\x00git push\x00x\x00\x00")
	if gen != "abc" || len(m) != 2 || m["gp"] != "git push" || m["x"] != "" {
		t.Errorf("%q %v", gen, m)
	}
	for _, empty := range []string{"", "\x00", "abc\x00"} {
		if _, m := ParseState(empty); m == nil || len(m) != 0 {
			t.Errorf("ParseState(%q) map = %v, want empty non-nil", empty, m)
		}
	}
	if _, m := ParseState("g\x00multi\nline\x00cmd\nwith newline\x00"); m["multi\nline"] != "cmd\nwith newline" {
		t.Errorf("newlines not preserved: %v", m)
	}
}

func TestCompare(t *testing.T) {
	store := map[string]string{"gp": "git pull", "k": "kubectl", "gone": "echo"}
	snap := map[string]string{"gp": "git push", "k": "kubectl", "other": "x"}
	got := Compare(store, snap)
	if len(got) != 2 || got["gp"] != "git push" || got["gone"] != "" {
		t.Errorf("%v", got)
	}
}

func TestStaleSnapshotIsSkipped(t *testing.T) {
	c := byName(Run(Input{RC: render.Block(), Overridden: map[string]string{}, SnapshotStale: true}))[CheckRedefined]
	if c.Status != Skipped || !strings.Contains(c.Detail, "new terminal") {
		t.Errorf("%+v", c)
	}
}
