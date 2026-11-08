package alias

import (
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	good := []string{"gp", "k", "git-pull", "a.b", "_x", "g@2", "ls+", "A1"}
	bad := []string{"", "-x", "a b", "a=b", "a/b", "a$b", "a'b", ".", "a\tb", "x" + string(make([]byte, 64))}
	for _, n := range good {
		if err := ValidateName(n); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", n, err)
		}
	}
	for _, n := range bad {
		if err := ValidateName(n); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", n)
		}
	}
}

func TestValidateGroup(t *testing.T) {
	for _, g := range []string{"git", "k8s", "my-tools", "a_b"} {
		if err := ValidateGroup(g); err != nil {
			t.Errorf("ValidateGroup(%q) = %v", g, err)
		}
	}
	for _, g := range []string{"", "Git", "-x", "a b", "version", "_pinned", "a.b"} {
		if err := ValidateGroup(g); err == nil {
			t.Errorf("ValidateGroup(%q) = nil, want error", g)
		}
	}
}

func TestValidateCommand(t *testing.T) {
	if err := ValidateCommand(""); err == nil {
		t.Error("empty command accepted")
	}
	if err := ValidateCommand("   "); err == nil {
		t.Error("blank command accepted")
	}
	if err := ValidateCommand("echo 'hi' && ls\nfoo"); err != nil {
		t.Errorf("complex command rejected: %v", err)
	}
}

func TestValidateNameRejectsWordsTheShellBlockRelaysOn(t *testing.T) {
	for _, n := range []string{"ali", "alias", "unalias", "builtin", "command", "source", "local", "return", "for", "do", "done", "unset"} {
		err := ValidateName(n)
		if err == nil || !strings.Contains(err.Error(), "reserved") {
			t.Errorf("ValidateName(%q) = %v, want reserved error", n, err)
		}
	}
}
