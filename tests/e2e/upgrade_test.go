package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// blockV1 is the .zshrc block exactly as released before feature 002.
const blockV1 = `# >>> ali >>>
_ali_dir="${ALI_CONFIG_DIR:-${XDG_CONFIG_HOME:-$HOME/.config}/ali}"
[ -f "$_ali_dir/aliases.zsh" ] && \source "$_ali_dir/aliases.zsh"
ali() { \command ali "$@"; \local rc=$?; [ -f "$_ali_dir/aliases.zsh" ] && \source "$_ali_dir/aliases.zsh"; \unset _n; \return $rc; }
# <<< ali <<<
`

func TestInitUpgradesAV1BlockInPlace(t *testing.T) {
	s := newSandbox(t)
	rc := filepath.Join(s.home, ".zshrc")
	before := omzRC + "\n"
	after := "\nalias late='echo late'\nexport AFTER=1\n"
	os.WriteFile(rc, []byte(before+blockV1+after), 0o644)

	if out := s.ali("init", "zsh"); !strings.Contains(out, "updated the ali block") {
		t.Fatalf("init output:\n%s", out)
	}
	got, _ := os.ReadFile(rc)
	if !strings.HasPrefix(string(got), before) || !strings.HasSuffix(string(got), after) {
		t.Errorf("content outside the block changed:\n%s", got)
	}
	if strings.Contains(string(got), blockV1) || !strings.Contains(string(got), "# ali block v2") {
		t.Errorf("block not upgraded:\n%s", got)
	}
	// the upgraded block brings the new behaviour into the shell
	if sh := s.zsh("ali add gp 'echo pulled'"); !strings.Contains(sh, "already an alias in this shell") {
		t.Errorf("upgraded block does not report shell definitions:\n%s", sh)
	}
	if out := s.ali("init", "zsh"); !strings.Contains(out, "already set up") {
		t.Errorf("second init:\n%s", out)
	}
	if backups, _ := filepath.Glob(rc + ".ali-backup-*"); len(backups) != 1 {
		t.Errorf("want exactly one backup, got %v", backups)
	}
}
