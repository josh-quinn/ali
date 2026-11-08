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
	if !strings.HasPrefix(out, Zsh([]alias.Alias{{Name: "gp", Command: "git pull", Group: "git"}})) || strings.Contains(out, "functions[") {
		t.Error("plain aliases must render as Zsh does, plus only the function cleanup")
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

func TestPlaceholderAliasesBecomeFunctions(t *testing.T) {
	in := expInput()
	plain, funcs := Split(in)
	for _, a := range plain {
		if a.Name == "kgp" {
			t.Error("kgp has a placeholder and must not be a plain alias")
		}
	}
	if len(funcs) != 1 || funcs[0].Name != "kgp" {
		t.Errorf("funcs = %v", funcs)
	}
	out := Render(in)
	if strings.Contains(out, "\\alias -- kgp=") || !strings.Contains(out, "functions[kgp]=") {
		t.Errorf("kgp not rendered as a function:\n%s", out)
	}
}

func TestFunctionsAreRemovedWhenNoLongerNeeded(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not installed")
	}
	dir := t.TempDir()
	os.WriteFile(dir+"/1.zsh", []byte(Render(expInput())), 0o600)
	without := expInput()
	without.Aliases = without.Aliases[:2]
	os.WriteFile(dir+"/2.zsh", []byte(Render(without)), 0o600)
	out, _ := exec.Command(zsh, "-f", "-c", "source "+dir+"/1.zsh\nprint -r -- \"before=${+functions[kgp]}\"\nsource "+dir+"/2.zsh\nprint -r -- \"after=${+functions[kgp]}\"").CombinedOutput()
	if !strings.Contains(string(out), "before=1") || !strings.Contains(string(out), "after=0") {
		t.Errorf("%s", out)
	}
}

func TestRuntimeModeRendersFunctionsInsteadOfWidgets(t *testing.T) {
	in := expInput()
	in.Runtime = true
	out := Render(in)
	// typing never rewrites the line: no Enter or Space widgets (Tab only suggests)
	if strings.Contains(out, "_ali_accept") || strings.Contains(out, "_ali_space") {
		t.Errorf("runtime mode must not rewrite the command line:\n%s", out)
	}
	if !strings.Contains(out, "_ali_tab") {
		t.Error("Tab completion should still be available in runtime mode")
	}
	for _, name := range []string{"k", "kgp", "gp"} { // gp: shared sub-aliases make git known
		if !strings.Contains(out, "functions["+name+"]=") || strings.Contains(out, "\\alias -- "+name+"=") {
			t.Errorf("%s should be a function in runtime mode:\n%s", name, out)
		}
	}
	plain, _ := Split(noSharedRuntime())
	for _, a := range plain {
		if a.Name == "k" {
			t.Error("k expands (kubectl has sub-aliases) and must be a function")
		}
	}
}

func noSharedRuntime() Input { in := noShared(); in.Runtime = true; return in }
