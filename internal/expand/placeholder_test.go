package expand

import (
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/model"
)

func vdefs() *Defs {
	d := defs()
	d.Aliases["kgp"] = "kubectl get pods -n {{ns}}"
	d.Subs["kubectl"]["n"] = "-n {{ns}}"
	d.Subs["kubectl"]["c"] = "--context={{ctx}}"
	d.Vars = map[string]model.Variable{
		"ns":  {Values: map[string]string{"one": "namespace-1", "sys": "kube-system", "dev": "dev-ns"}, Default: "one"},
		"ctx": {Values: map[string]string{"dev": "dev-cluster", "prod": "prod-cluster"}},
	}
	d.Envs = map[string]model.Environment{
		"dev":  {Set: map[string]string{"ns": "dev", "ctx": "dev"}},
		"prod": {Protected: true, Set: map[string]string{"ns": "production", "ctx": "prod"}},
	}
	return d
}

func full(t *testing.T, line string, opt Options) (Result, error) {
	t.Helper()
	opt.Mode = Full
	return Expand(line, vdefs(), opt)
}

func TestPlaceholderResolution(t *testing.T) {
	cases := []struct {
		in, want string
		opt      Options
	}{
		{"k g p n @sys", "kubectl get pods -n kube-system", Options{}},
		{"k g p n", "kubectl get pods -n namespace-1", Options{}},         // default
		{"k g p n @ns=my-test", "kubectl get pods -n my-test", Options{}}, // literal
		{"k g p n @ns=sys", "kubectl get pods -n kube-system", Options{}}, // key wins
		{"kgp @sys", "kubectl get pods -n kube-system", Options{}},        // alias with placeholder
		{"k g p n @sys c @prod", "kubectl get pods -n kube-system --context=prod-cluster", Options{}},
		{"k g p n @sys n", "kubectl get pods -n kube-system -n kube-system", Options{}},               // same value twice
		{"k g p n", "kubectl get pods -n dev-ns", Options{TerminalEnv: "dev"}},                        // terminal env (key)
		{"k g p n", "kubectl get pods -n dev-ns", Options{GlobalEnv: "dev"}},                          // global env
		{"k g p n", "kubectl get pods -n production", Options{TerminalEnv: "prod", GlobalEnv: "dev"}}, // terminal over global
		{"k g p n @@dev", "kubectl get pods -n dev-ns", Options{TerminalEnv: "prod"}},                 // @@ over terminal
		{"k g p n @sys", "kubectl get pods -n kube-system", Options{TerminalEnv: "dev"}},              // explicit over env
		{"k g p n @ns='a b; rm -rf ~'", "kubectl get pods -n 'a b; rm -rf ~'", Options{}},             // quoted value: one literal word
		{`k g p n @ns="x y"`, "kubectl get pods -n 'x y'", Options{}},
		{`k g p n @ns=a\ b`, "kubectl get pods -n 'a b'", Options{}},
		{`k g p n @ns="$HOME"`, `kubectl get pods -n namespace-1 @ns="$HOME"`, Options{}}, // expansions are never evaluated by ali
		{"k g p n @ns=a;b", "kubectl get pods -n a;b", Options{}},                         // ; ends the command: @ns=a, then command b
	}
	for _, c := range cases {
		res, err := full(t, c.in, c.opt)
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if res.Line != c.want {
			t.Errorf("%q => %q, want %q", c.in, res.Line, c.want)
		}
	}
}

func TestHostileValuesAreQuotedWhenSubstituted(t *testing.T) {
	d := vdefs()
	v := d.Vars["ns"]
	v.Values["evil"] = "x; rm -rf ~ $(id)"
	res, err := Expand("k g p n @evil", d, Options{Mode: Full})
	if err != nil || res.Line != `kubectl get pods -n 'x; rm -rf ~ $(id)'` {
		t.Errorf("%q %v", res.Line, err)
	}
}

func TestUnrelatedAtWordsAreLeftAlone(t *testing.T) {
	for _, in := range []string{"npm i @types/node", "curl -d @body.json x", "k g p @sys", "k g p n @nope"} {
		res, err := full(t, in, Options{})
		if err != nil {
			t.Errorf("%q: %v", in, err)
		}
		if strings.Contains(in, "@types") || strings.Contains(in, "@body") {
			if res.Line != in {
				t.Errorf("%q changed to %q", in, res.Line)
			}
		}
	}
	// k g p @sys: no placeholder in the line, so @sys stays
	if res, _ := full(t, "k g p @sys", Options{}); res.Line != "kubectl get pods @sys" {
		t.Errorf("got %q", res.Line)
	}
	// unknown key with a placeholder present: @nope stays, default used
	if res, _ := full(t, "k g p n @nope", Options{}); res.Line != "kubectl get pods -n namespace-1 @nope" {
		t.Errorf("got %q", res.Line)
	}
}

func TestAmbiguousAndMissingValues(t *testing.T) {
	// "dev" is a key of both ns and ctx
	_, err := full(t, "k g p n c @dev", Options{})
	if err == nil || !strings.Contains(err.Error(), "@ns=dev") || !strings.Contains(err.Error(), "@ctx=dev") {
		t.Errorf("want ambiguity error naming both, got %v", err)
	}
	// ctx has no default and no env
	_, err = full(t, "k g p c", Options{})
	if err == nil || !strings.Contains(err.Error(), "ctx") {
		t.Errorf("want missing-value error, got %v", err)
	}
	// unknown @@env with placeholders present
	_, err = full(t, "k g p n @@nowhere", Options{})
	if err == nil || !strings.Contains(err.Error(), "nowhere") {
		t.Errorf("want unknown environment error, got %v", err)
	}
}

func TestProtectedAndStaleEnvironments(t *testing.T) {
	res, err := full(t, "k g p n", Options{TerminalEnv: "prod"})
	if err != nil || !res.Protected || res.Env != "prod" {
		t.Errorf("protected not reported: %+v %v", res, err)
	}
	if res, _ := full(t, "k g p n @sys", Options{TerminalEnv: "prod"}); res.Protected {
		t.Error("explicit value must not count as coming from the protected environment")
	}
	res, err = full(t, "k g p n", Options{TerminalEnv: "deleted"})
	if err != nil || res.Line != "kubectl get pods -n namespace-1" || len(res.Warnings) == 0 {
		t.Errorf("stale ALI_ENV should warn and fall back: %+v %v", res, err)
	}
}

func TestPartialModeLeavesPlaceholders(t *testing.T) {
	res, err := Expand("k g p n", vdefs(), Options{Mode: Partial})
	if err != nil || res.Line != "kubectl get pods -n {{ns}}" {
		t.Errorf("%q %v", res.Line, err)
	}
	res, _ = Expand("kgp", vdefs(), Options{Mode: Partial})
	if res.Line != "kubectl get pods -n {{ns}}" {
		t.Errorf("alias with placeholder should expand on Space: %q", res.Line)
	}
}

func TestUndefinedPlaceholdersAreLiteral(t *testing.T) {
	d := vdefs()
	d.Subs["kubectl"]["t"] = "-o go-template={{end}}"
	res, err := Expand("k g p t", d, Options{Mode: Full})
	if err != nil || res.Line != "kubectl get pods -o go-template={{end}}" {
		t.Errorf("%q %v", res.Line, err)
	}
}
