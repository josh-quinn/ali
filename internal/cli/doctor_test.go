package cli

import (
	"os"

	"path/filepath"
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/render"
)

// state builds what the wrapper would send: the generation of the current store
// followed by a snapshot of the shell's aliases (store aliases plus overrides).
func state(e *env, overrides map[string]string) string {
	d, _ := e.st.Load()
	s := render.Gen(d.All()) + "\x00"
	for _, a := range d.All() {
		cmd := a.Command
		if o, ok := overrides[a.Name]; ok {
			cmd = o
		}
		s += a.Name + "\x00" + cmd + "\x00"
	}
	return s
}

func TestDoctorHealthyThroughWrapper(t *testing.T) {
	e := newEnv(t)
	e.must("init", "zsh")
	e.must("add", "gp", "git pull")
	out, _, err := e.runIn(state(e, nil), "doctor", "--shell-state", "-")
	if err != nil {
		t.Fatalf("healthy setup failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "FAIL") || strings.Count(out, "ok ") < 4 {
		t.Errorf("output:\n%s", out)
	}
}

func TestDoctorOutsideWrapperSkipsLiveCheck(t *testing.T) {
	e := newEnv(t)
	e.must("init", "zsh")
	out, _, err := e.run("doctor")
	if err != nil {
		t.Fatalf("skipped checks alone must not fail: %v", err)
	}
	if !strings.Contains(out, "skipped") || !strings.Contains(out, "from your shell") {
		t.Errorf("output:\n%s", out)
	}
}

func TestDoctorReportsProblemsAndExitsNonZero(t *testing.T) {
	e := newEnv(t)
	rc := filepath.Join(e.dir, ".zshrc")
	e.must("init", "zsh")
	b, _ := os.ReadFile(rc)
	os.WriteFile(rc, append(b, []byte("source $ZSH/oh-my-zsh.sh\n")...), 0o644)

	e.must("add", "gp", "git pull")
	out, _, err := e.runIn(state(e, map[string]string{"gp": "git push"}), "doctor", "--shell-state", "-")
	if err == nil {
		t.Fatal("want failure exit")
	}
	for _, want := range []string{"FAIL", "block loads after Oh My Zsh", "fix:", "gp (now: git push)"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorWithoutBlock(t *testing.T) {
	e := newEnv(t)
	out, _, err := e.run("doctor")
	if err == nil || !strings.Contains(out, "ali init zsh") {
		t.Errorf("err=%v out=\n%s", err, out)
	}
}

func TestDoctorSanitisesShellState(t *testing.T) {
	e := newEnv(t)
	e.must("init", "zsh")
	e.must("add", "gp", "git pull")
	out, _, _ := e.runIn(state(e, map[string]string{"gp": "\x1b[2J"}), "doctor", "--shell-state", "-")
	if strings.ContainsRune(out, '\x1b') {
		t.Errorf("escape leaked: %q", out)
	}
}

func TestDoctorSkipsWhenStoreChangedSinceTerminalStarted(t *testing.T) {
	e := newEnv(t)
	e.must("init", "zsh")
	e.must("add", "gp", "git pull")
	snap := state(e, nil)
	e.must("add", "new", "echo new") // added after the terminal started
	out, _, err := e.runIn(snap, "doctor", "--shell-state", "-")
	if err != nil || !strings.Contains(out, "skipped aliases not redefined") || strings.Contains(out, "new (now") {
		t.Errorf("err=%v\n%s", err, out)
	}
}
