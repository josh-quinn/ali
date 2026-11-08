package expand

import (
	"os/exec"
	"testing"
)

func TestQuoteValueLeavesSafeValuesReadable(t *testing.T) {
	for _, v := range []string{"kube-system", "namespace-1", "a.b/c:d", "x=y", "50%", "u@host", "-n"} {
		if got := QuoteValue(v); got != v {
			t.Errorf("QuoteValue(%q) = %q, want unchanged", v, got)
		}
	}
	for _, v := range []string{"=ls", "~/x", "a b", "{a,b}", "*", "a;b"} {
		if QuoteValue(v) == v {
			t.Errorf("QuoteValue(%q) left unquoted", v)
		}
	}
}

// SC-004: every hostile value reaches the command as exactly one literal argument.
func TestHostileValuesArriveAsOneLiteralArgument(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not installed")
	}
	hostile := []string{
		"a b; rm -rf ~", "$(touch /tmp/ali-pwned)", "`id`", "it's", `"quoted"`, "a\nb",
		"$HOME", "${PATH}", "!!", "*", "~", "=ls", "a|b", "a&b", "a>b", "\\", "{{ns}}", "@sys", "--", "-rf",
		"\t", "é ✓", "a\\'b",
	}
	for _, v := range hostile {
		script := "setopt nobanghist; args=( " + QuoteValue(v) + " ); print -rn -- \"${#args}\"$'\\0'\"$args[1]\""
		out, err := exec.Command(zsh, "-f", "-c", script).CombinedOutput()
		if err != nil {
			t.Errorf("%q: zsh failed: %v %q", v, err, out)
			continue
		}
		if want := "1\x00" + v; string(out) != want {
			t.Errorf("%q arrived as %q", v, out)
		}
	}
}
