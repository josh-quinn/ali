package expand

import (
	"testing"

	"github.com/josh-quinn/ali/internal/model"
)

func defs() *Defs {
	return &Defs{
		Aliases: map[string]string{"k": "kubectl", "h": "helm", "gp": "git pull", "sk": "sudo -E kubectl"},
		Subs: map[string]map[string]string{
			"kubectl": {"g": "get", "p": "pods", "d": "describe", "o": "-o yaml"},
			"helm":    {"g": "get values"},
		},
		Shared: map[string]string{"o": "-o wide", "w": "--watch"},
		Vars:   map[string]model.Variable{},
		Envs:   map[string]model.Environment{},
	}
}

func TestSubAliasExpansion(t *testing.T) {
	cases := []struct{ in, want string }{
		{"k g p", "kubectl get pods"},
		{"kubectl g p", "kubectl get pods"},           // tool typed directly
		{"k get p", "kubectl get pods"},               // anywhere after the tool (ADR 0031)
		{"k logs mypod p", "kubectl logs mypod pods"}, // even after literal words
		{`k g \p`, `kubectl get \p`},                  // backslash keeps a word literal
		{`k logs \p`, `k logs \p`},                    // nothing to expand: k is left for zsh
		{"k g 'p' \"p\"", "kubectl get 'p' \"p\""},    // quoted words never expand
		{"k exec pod -- g p", "k exec pod -- g p"},    // bare -- ends expansion
		{"k d pod -- g p", "kubectl describe pod -- g p"},
		{"k g p o", "kubectl get pods -o yaml"}, // tool's own wins over shared
		{"k g p w", "kubectl get pods --watch"}, // shared applies to known tools
		{"h g", "helm get values"},              // same name, different tool
		{"k g | h g; kubectl d p", "kubectl get | helm get values; kubectl describe pods"},
		{"sudo k g p", "sudo kubectl get pods"}, // wrappers are looked through
		{"FOO=1 time k g", "FOO=1 time kubectl get"},
		{"sk g p", "sudo -E kubectl get pods"},        // alias that starts with a wrapper
		{"  k   g   p  ", "  kubectl   get   pods  "}, // spacing kept
	}
	for _, c := range cases {
		res, err := Expand(c.in, defs(), Options{Mode: Full})
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if res.Line != c.want {
			t.Errorf("Expand(%q) = %q, want %q", c.in, res.Line, c.want)
		}
		if res.Changed != (c.in != c.want) {
			t.Errorf("%q: Changed = %v", c.in, res.Changed)
		}
	}
}

func TestLinesWithoutExpansionAreUntouched(t *testing.T) {
	for _, in := range []string{
		"k",             // plain alias alone: zsh expands it as usual (FR-017)
		"k logs mypod",  // alias but no sub-alias used
		"gp",            // alias for a tool without sub-aliases
		"ls -la g p",    // unknown tool: no sub-aliases, shared ones do not apply
		"cp o x",        // shared sub-aliases only apply to tools ali knows
		"'k' g p",       // quoted command word
		"$K g p",        // not a plain command word
		"echo $(k g p)", // inside a substitution: left alone
		"",
	} {
		res, err := Expand(in, defs(), Options{Mode: Full})
		if err != nil || res.Changed || res.Line != in {
			t.Errorf("Expand(%q) = %q changed=%v err=%v", in, res.Line, res.Changed, err)
		}
	}
}

func TestExpansionIsNotRecursive(t *testing.T) {
	d := defs()
	d.Subs["kubectl"]["x"] = "g p" // expansion text containing sub-alias names
	res, _ := Expand("k x", d, Options{Mode: Full})
	if res.Line != "kubectl g p" {
		t.Errorf("got %q", res.Line)
	}
}

func TestToolOf(t *testing.T) {
	d := defs()
	for in, want := range map[string]string{"k g": "kubectl", "sudo -u root kubectl x": "kubectl", "env -u X h g": "helm", "/usr/local/bin/kubectl g": "kubectl", "ls": "ls", "$x": ""} {
		cmds := Tokenize(in)
		if got := d.toolOf(cmds[0]).tool; got != want {
			t.Errorf("toolOf(%q) = %q, want %q", in, got, want)
		}
	}
}
