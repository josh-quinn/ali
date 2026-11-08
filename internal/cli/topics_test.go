package cli

import (
	"strings"
	"testing"
)

func TestHelpFilesTopic(t *testing.T) {
	e := newEnv(t)
	out := e.must("help", "files")
	for _, want := range []string{"aliases.toml", "aliases.zsh", "aliases.toml.lock", ".ali-backup-", "ALI_CONFIG_DIR", "XDG_CONFIG_HOME", "ZDOTDIR", "0600"} {
		if !strings.Contains(out, want) {
			t.Errorf("help files does not mention %q:\n%s", want, out)
		}
	}
	if root := e.must("--help"); !strings.Contains(root, "files") {
		t.Errorf("root help does not list the files topic:\n%s", root)
	}
}
