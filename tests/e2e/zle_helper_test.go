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
all=
# Segments separated by \x1e alternate between keys to type and text to wait
# for on screen ("k g p c\r\x1eno value for\x1e\x15..."), so a test never
# sends its next key while a widget may still be running.
segs=("${(@ps:\x1e:)keys}")
for (( s = 1; s <= $#segs; s++ )); do
  if (( s % 2 == 0 )); then
    zpty -r z chunk "*${(b)segs[s]}*" && all+=$chunk
    continue
  fi
  seg=${segs[s]}
  # type one key at a time, like a person; a burst would look like a paste
  for (( i = 1; i <= $#seg; i++ )); do zpty -w -n z "${seg[i]}"; zselect -t 1; done
done
zpty -w -n z $'\rprint -r -- __ALI_DO""NE__\r'
zpty -r z out '*__ALI_DONE__*' && print -rn -- "$all$out"
zpty -d z
`

// CSI sequences, OSC sequences (terminal titles set by themes such as Oh My
// Zsh's), keypad modes and carriage returns.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;?]*[A-Za-z]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[=>]|\r`)

// zle types keys (use "\r" for Enter, "\t" for Tab) into a fresh interactive
// shell in the sandbox and returns the screen output with escapes stripped.
// waitFor(text) inside keys pauses typing until text has appeared on screen.
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

// waitFor makes the driver pause typing until text appears on screen. Use it
// before a key that must reach the line editor after a widget has finished.
func waitFor(text string) string { return "\x1e" + text + "\x1e" }

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

func TestZleDriverWaitsForText(t *testing.T) {
	s := newSandbox(t)
	got := s.zle("echo first\r" + waitFor("first") + "echo second\r")
	if !hasLine(got, "first") || !hasLine(got, "second") {
		t.Errorf("output around a wait is missing:\n%q", got)
	}
}
