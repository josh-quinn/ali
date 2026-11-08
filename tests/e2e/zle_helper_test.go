package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// zleDriver runs an interactive zsh inside zsh/zpty so the line editor is
// active (research R10: ZLE is off when stdin is not a terminal). It types the
// given keys, then a marker command, and returns everything the terminal showed.
const zleDriver = `zmodload zsh/zpty || { print -r -- "NO-ZPTY"; exit 2 }
zpty -b z zsh -i -d
zpty -w -n z $'PROMPT=""\r'
sleep 0.2
zpty -r -t z _junk
zmodload zsh/zselect
keys="$(<$1)"
# type one key at a time, like a person; a burst would look like a paste
for (( i = 1; i <= $#keys; i++ )); do zpty -w -n z "${keys[i]}"; zselect -t 1; done
zpty -w -n z $'\rprint -r -- __ALI_DO""NE__\r'
zpty -r z out '*__ALI_DONE__*' && print -rn -- "$out"
zpty -d z
`

// CSI sequences, OSC sequences (terminal titles set by themes such as Oh My
// Zsh's), keypad modes and carriage returns.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[=>]|\r`)

// zle types keys (use "\r" for Enter, "\t" for Tab) into a fresh interactive
// shell in the sandbox and returns the screen output with escapes stripped.
func (s *sandbox) zle(keys string) string {
	s.t.Helper()
	dir := s.t.TempDir()
	drv, kf := filepath.Join(dir, "drive.zsh"), filepath.Join(dir, "keys")
	os.WriteFile(drv, []byte(zleDriver), 0o600)
	os.WriteFile(kf, []byte(keys), 0o600)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "zsh", "-f", drv, kf)
	cmd.Env = append(s.env(), "TERM=xterm")
	out, err := cmd.CombinedOutput()
	if strings.Contains(string(out), "NO-ZPTY") {
		s.t.Skip("zsh/zpty not available")
	}
	if ctx.Err() != nil {
		s.t.Fatalf("zle driver timed out; output so far:\n%s", out)
	}
	if err != nil {
		s.t.Fatalf("zle driver: %v\n%s", err, out)
	}
	return ansiRe.ReplaceAllString(string(out), "")
}

func TestZleDriverSelfTest(t *testing.T) {
	s := newSandbox(t)
	got := s.zle("echo hel\"\"lo\rfc -ln -1 -1\r")
	if !hasLine(got, "hello") {
		t.Errorf("command output missing:\n%q", got)
	}
	// an exact line (not the redrawn typed text) can only come from fc's output
	if !hasLine(got, `echo hel""lo`) {
		t.Errorf("history output missing:\n%q", got)
	}
}
