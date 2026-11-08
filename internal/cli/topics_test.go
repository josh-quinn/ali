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

func TestSyntaxHelpTopics(t *testing.T) {
	e := newEnv(t)
	cases := map[string][]string{
		"expansion":    {`\p`, "--", "ali sub add", "ali config set expand runtime", "visible"},
		"placeholders": {"{{ns}}", "@sys", "@ns=", "default", "quoted"},
		"environments": {"@@dev", "--global", "protected", "Enter again", "[ali:", "ali_prompt_info"},
	}
	for topic, wants := range cases {
		out := e.must("help", topic)
		for _, w := range wants {
			if !strings.Contains(out, w) {
				t.Errorf("help %s missing %q", topic, w)
			}
		}
	}
}
