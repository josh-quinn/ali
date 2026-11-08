package e2e

import (
	"strings"
	"testing"
)

func TestRuntimeModeKeepsTheTypedLine(t *testing.T) {
	s := placeholderSandbox(t)
	s.ali("config", "set", "expand", "runtime")
	got := s.zle("k g p n @sys\rfc -ln -1 -1\rkubectl g p\r")
	if !strings.Contains(got, "[get][pods][-n][kube-system]") {
		t.Errorf("command did not run expanded:\n%s", got)
	}
	if !hasLine(got, "k g p n @sys") {
		t.Errorf("history should keep the line as typed:\n%s", got)
	}
	if !strings.Contains(got, "[g][p]") {
		t.Errorf("a tool-first line must not expand in run-time mode:\n%s", got)
	}
	if out := s.ali("expand", "k g p n @sys"); strings.TrimSpace(out) != "kubectl get pods -n kube-system" {
		t.Errorf("ali expand: %q", out)
	}
}
