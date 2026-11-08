package model

import (
	"reflect"
	"testing"
)

func TestNameRules(t *testing.T) {
	type rule struct {
		name string
		fn   func(string) error
		good []string
		bad  []string
	}
	rules := []rule{
		{"sub", ValidateSubName, []string{"g", "p", "gp2", "o.w", "x-y"}, []string{"", "@x", `\p`, "-x", "a b", "a=b", "{{x}}"}},
		{"tool", ValidateTool, []string{"kubectl", "helm", "k9s", "docker-compose"}, []string{"", "Kubectl", "a b", "_x", "--"}},
		{"var", ValidateVarName, []string{"ns", "ctx", "my_var2"}, []string{"", "Ns", "2x", "a-b", "env", "a b"}},
		{"key", ValidateKey, []string{"one", "sys", "kube-system", "1", "a_b"}, []string{"", "One", "a b", "a=b", "-x"}},
		{"env", ValidateEnvName, []string{"dev", "prod", "eu-west_1"}, []string{"", "Dev", "1x", "a b", "a=b"}},
	}
	for _, r := range rules {
		for _, n := range r.good {
			if err := r.fn(n); err != nil {
				t.Errorf("%s %q rejected: %v", r.name, n, err)
			}
		}
		for _, n := range r.bad {
			if err := r.fn(n); err == nil {
				t.Errorf("%s %q accepted", r.name, n)
			}
		}
	}
}

func TestValidateValue(t *testing.T) {
	if ValidateValue("") == nil || ValidateValue("a\x00b") == nil {
		t.Error("empty or NUL value accepted")
	}
	if err := ValidateValue("a b; $(rm -rf ~)"); err != nil {
		t.Errorf("values are data and may contain anything else: %v", err)
	}
}

func TestPlaceholderNames(t *testing.T) {
	got := PlaceholderNames("kubectl -n {{ns}} --context={{ctx}} {{ns}} {{.Name}} {{ end }} {{Bad}} {{2x}}")
	if want := []string{"ns", "ctx"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestVariableResolveKey(t *testing.T) {
	v := Variable{Values: map[string]string{"one": "namespace-1"}, Default: "one"}
	if got, ok := v.Lookup("one"); !ok || got != "namespace-1" {
		t.Errorf("key lookup: %q %v", got, ok)
	}
	if _, ok := v.Lookup("nope"); ok {
		t.Error("unknown key found")
	}
	if err := v.Validate(); err != nil {
		t.Error(err)
	}
	if err := (Variable{Values: map[string]string{"one": "x"}, Default: "two"}).Validate(); err == nil {
		t.Error("default must be a key")
	}
}
