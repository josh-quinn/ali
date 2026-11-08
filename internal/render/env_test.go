package render

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/alias"
)

// runBlockWith is runBlock with a custom aliases.zsh input and global-env.
func runBlockWith(t *testing.T, in Input, global, script string) string {
	t.Helper()
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not installed")
	}
	dir := t.TempDir()
	bin, cfg := dir+"/bin", dir+"/cfg"
	os.Mkdir(bin, 0o755)
	os.Mkdir(cfg, 0o700)
	// fake ali: "fail" as 3rd arg exits 1; otherwise prints what it got
	os.WriteFile(bin+"/ali", []byte("#!/bin/sh\n[ \"$3\" = fail ] && exit 1\necho \"wrapper=${ALI_WRAPPER-} args=$*\"\n"), 0o755)
	os.WriteFile(cfg+"/aliases.zsh", []byte(Render(in)), 0o600)
	if global != "" {
		os.WriteFile(cfg+"/global-env", []byte(global), 0o600)
	}
	os.WriteFile(dir+"/block.zsh", []byte(Block()), 0o600)
	cmd := exec.Command(zsh, "-f", "-c", "source "+dir+"/block.zsh\n"+script)
	cmd.Env = []string{"PATH=" + bin + ":/usr/bin:/bin", "ALI_CONFIG_DIR=" + cfg, "HOME=" + dir}
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func envInput() Input {
	return Input{Aliases: []alias.Alias{{Name: "k", Command: "kubectl", Group: "kubectl"}}, Envs: []string{"dev", "prod"}, Protected: []string{"prod"}}
}

func TestWrapperSetsTheTerminalEnvironment(t *testing.T) {
	out := runBlockWith(t, envInput(), "", `ali env use prod; print -r -- "1:${ALI_ENV-unset}"
ali env use dev --global; print -r -- "2:${ALI_ENV-unset}"
ali env use prod; ali env off; print -r -- "3:${ALI_ENV-unset}"
ali env use prod; ali env use fail; print -r -- "4:${ALI_ENV-unset}"
ali env use '$(touch PWNED)'; print -r -- "5:${ALI_ENV-unset}"`)
	for _, want := range []string{"wrapper=1 args=env use prod", "1:prod", "2:unset", "3:unset", "4:prod", "5:prod"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestPromptShowsActiveEnvironment(t *testing.T) {
	cases := []struct {
		name, global, script, want string
	}{
		{"terminal", "", "ALI_ENV=dev; _ali_prompt; print -r -- \"[$RPROMPT]\"", "[%F{cyan}[ali:dev]%f]"},
		{"protected", "", "ALI_ENV=prod; _ali_prompt; print -r -- \"[$RPROMPT]\"", "[%F{red}[ali:prod]%f]"},
		{"global", "dev\n", "_ali_prompt; print -r -- \"[$RPROMPT]\"", "[%F{cyan}[ali:dev]%f]"},
		{"none", "", "_ali_prompt; print -r -- \"[$RPROMPT]\"", "[]"},
		{"keeps theme", "", "RPROMPT='%~'; ALI_ENV=dev; _ali_prompt; _ali_prompt; print -r -- \"[$RPROMPT]\"", "[%F{cyan}[ali:dev]%f%~]"},
		{"switch", "", "RPROMPT='x'; ALI_ENV=dev; _ali_prompt; ALI_ENV=prod; _ali_prompt; unset ALI_ENV; _ali_prompt; print -r -- \"[$RPROMPT]\"", "[x]"},
		{"hostile global", "$(touch PWNED)\n", "_ali_prompt; print -r -- \"[$RPROMPT]\"", "[]"},
		{"hostile terminal", "", "ALI_ENV='%{$(id)%}'; _ali_prompt; print -r -- \"[$RPROMPT]\"", "[]"},
		{"info function", "", "ALI_ENV=dev; _ali_prompt; print -r -- \"<$(ali_prompt_info)>\"", "<%F{cyan}[ali:dev]%f>"},
	}
	for _, c := range cases {
		out := runBlockWith(t, envInput(), c.global, c.script)
		if !strings.Contains(out, c.want) {
			t.Errorf("%s: want %q in %q", c.name, c.want, out)
		}
	}
	off := envInput()
	off.PromptOff = true
	if out := runBlockWith(t, off, "", "RPROMPT=x; ALI_ENV=dev; _ali_prompt; print -r -- \"[$RPROMPT]\""); !strings.Contains(out, "[x]") {
		t.Errorf("prompt=off still shows the label: %q", out)
	}
}

func TestBlockVersion3(t *testing.T) {
	if BlockVersion != 3 || !strings.Contains(Block(), "_ali_prompt") {
		t.Errorf("block not bumped for environments:\n%s", Block())
	}
}
