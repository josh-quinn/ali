package e2e

import (
	"strings"
	"testing"
)

func envSandbox(t *testing.T) *sandbox {
	s := placeholderSandbox(t)
	s.ali("env", "add", "dev", "ns=sys")
	s.ali("env", "add", "prod", "ns=production", "--protect")
	return s
}

func TestEnvironmentIsPerTerminal(t *testing.T) {
	s := envSandbox(t)
	a := s.zle("ali env use dev\rk g p n\r")
	if !strings.Contains(a, "[get][pods][-n][kube-system]") {
		t.Errorf("terminal A did not use dev:\n%s", a)
	}
	b := s.zle("k g p n\r") // a new terminal: no choice of its own, no global
	if !strings.Contains(b, "[get][pods][-n][namespace-1]") {
		t.Errorf("terminal B was affected by A:\n%s", b)
	}
}

func TestGlobalEnvironmentAndTerminalOverride(t *testing.T) {
	s := envSandbox(t)
	s.zle("ali env use dev --global\r")
	got := s.zle("k g p n\rali env use prod\rk g p n @@dev\r")
	if !strings.Contains(got, "[get][pods][-n][kube-system]") {
		t.Errorf("new terminal did not follow the global choice:\n%s", got)
	}
	if strings.Count(got, "[kube-system]") != 2 {
		t.Errorf("@@dev should override the terminal's prod for one line:\n%s", got)
	}
}

func TestPromptShowsTheEnvironment(t *testing.T) {
	s := envSandbox(t)
	got := s.zle("ali env use dev\rtrue\r")
	if !strings.Contains(got, "[ali:dev]") {
		t.Errorf("prompt label missing:\n%q", got)
	}
}

func TestProtectedEnvironmentNeedsASecondEnter(t *testing.T) {
	s := envSandbox(t)
	// Enter, wait for the protected notice, then Ctrl-U clears the kept line
	once := s.zle("ali env use prod\rk g p n\r" + waitFor("press Enter again") + "\x15echo cancelled\r")
	if strings.Contains(once, "[production]") {
		t.Errorf("ran after a single Enter:\n%s", once)
	}
	if !strings.Contains(once, "protected") || !hasLine(once, "cancelled") {
		t.Errorf("want a protected notice and a usable shell:\n%s", once)
	}
	twice := s.zle("ali env use prod\rk g p n\r" + waitFor("press Enter again") + "\r")
	if strings.Count(twice, "[get][pods][-n][production]") != 1 {
		t.Errorf("second Enter should run it exactly once:\n%s", twice)
	}
	// an explicit value does not come from prod, so no confirmation is needed
	explicit := s.zle("ali env use prod\rk g p n @sys\r")
	if !strings.Contains(explicit, "[get][pods][-n][kube-system]") {
		t.Errorf("explicit value should run at once:\n%s", explicit)
	}
}
