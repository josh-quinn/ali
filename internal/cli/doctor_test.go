package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorHealthyThroughWrapper(t *testing.T) {
	e := newEnv(t)
	e.must("init", "zsh")
	out, _, err := e.runIn("", "doctor", "--shell-state", "-")
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

	out, _, err := e.runIn("gp\x00git push\x00", "doctor", "--shell-state", "-")
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
	out, _, _ := e.runIn("gp\x00\x1b[2J\x00", "doctor", "--shell-state", "-")
	if strings.ContainsRune(out, '\x1b') {
		t.Errorf("escape leaked: %q", out)
	}
}
