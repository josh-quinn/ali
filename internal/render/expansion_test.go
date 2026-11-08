package render

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/alias"
)

func expInput() Input {
	return Input{
		Aliases: []alias.Alias{
			{Name: "k", Command: "kubectl", Group: "kubectl"},
			{Name: "gp", Command: "git pull", Group: "git"},
			{Name: "kgp", Command: "kubectl get pods -n {{ns}}", Group: "kubectl"},
		},
		Subs:   map[string]map[string]string{"kubectl": {"g": "get", "p": "pods"}},
		Shared: map[string]string{"o": "-o wide"},
		Vars:   []string{"ns"},
	}
}

func TestNoExpansionCodeWithoutSubAliasesOrPlaceholders(t *testing.T) {
	out := Render(Input{Aliases: []alias.Alias{{Name: "gp", Command: "git pull", Group: "git"}}})
	if strings.Contains(out, "_ali_trigger") || strings.Contains(out, "bindkey") || strings.Contains(out, "zle") {
		t.Errorf("plain-alias users must pay nothing for expansion:\n%s", out)
	}
	if out != Zsh([]alias.Alias{{Name: "gp", Command: "git pull", Group: "git"}}) {
		t.Error("Zsh and Render differ for plain aliases")
	}
}

func noShared() Input { in := expInput(); in.Shared = nil; return in }

func TestTriggerAndSubWordTables(t *testing.T) {
	out := Render(noShared())
	for _, want := range []string{
		"_ali_trigger=('k' 1 'kgp' 1 'kubectl' 1)", // tool, alias for it, placeholder alias
		"_ali_subw=('g' 1 'kgp' 1 'p' 1)",          // sub names and placeholder aliases
		"\\zle -N _ali_accept",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "'gp' 1") {
		t.Error("gp (git, no sub-aliases) must not trigger expansion")
	}
	// with a shared sub-alias every tool an alias runs is known, so gp triggers
	if shared := Render(expInput()); !strings.Contains(shared, "_ali_trigger=('git' 1 'gp' 1 'k' 1 'kgp' 1 'kubectl' 1)") || !strings.Contains(shared, "'o' 1") {
		t.Errorf("shared tables wrong:\n%s", shared)
	}
}

// The gate decides in zsh, without a process, whether a line may expand.
func TestGateInZsh(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not installed")
	}
	f := t.TempDir() + "/a.zsh"
	os.WriteFile(f, []byte(Render(noShared())), 0o600)
	script := "source " + f + "\nfor l in 'k g' 'sudo kubectl x' 'kgp @sys' 'ls -la' 'gp' 'echo kubectl'; do _ali_gate \"$l\" && print -r -- \"yes:$l\" || print -r -- \"no:$l\"; done"
	out, _ := exec.Command(zsh, "-f", "-c", script).CombinedOutput()
	for _, want := range []string{"yes:k g", "yes:sudo kubectl x", "yes:kgp @sys", "no:ls -la", "no:gp", "yes:echo kubectl"} {
		if !strings.Contains(string(out), want+"\n") {
			t.Errorf("gate: missing %q in\n%s", want, out)
		}
	}
}

func TestWidgetsAreWrappedOnceAcrossReloads(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not installed")
	}
	f := t.TempDir() + "/a.zsh"
	os.WriteFile(f, []byte(Render(expInput())), 0o600)
	// sourcing twice must not make _ali_orig_accept point at ali's own widget
	script := "zmodload zsh/zleparameter\nsource " + f + "\nsource " + f + "\nprint -r -- \"orig=${widgets[_ali_orig_accept]} accept=${widgets[accept-line]}\""
	out, _ := exec.Command(zsh, "-f", "-i", "-c", script).CombinedOutput()
	if !strings.Contains(string(out), "orig=builtin") || !strings.Contains(string(out), "accept=user:_ali_accept") {
		t.Errorf("widget wrapping wrong after reload:\n%s", out)
	}
}

func TestExpansionCodeIsHijackSafe(t *testing.T) {
	out := Render(expInput())
	for _, bare := range []string{"\nzle ", " zle ", "\nlocal ", " local ", "bindkey ", "\nreturn", " return"} {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(" "+line, bare) && !strings.Contains(line, "\\"+strings.TrimSpace(bare)) {
				t.Errorf("unescaped %q in generated line %q", strings.TrimSpace(bare), line)
			}
		}
	}
}
