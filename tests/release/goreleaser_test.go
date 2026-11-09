// Package release checks .goreleaser.yaml for mistakes that `goreleaser check`
// and snapshot builds cannot catch, because the fields are only evaluated when
// publishing. v0.1.0 and v0.1.1 shipped without their Homebrew cask because of
// two such mistakes.
package release

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// GoReleaser v2.20.0, internal/tmpl/tmpl.go: token fields must be exactly one
// environment variable reference.
var envOnlyRe = regexp.MustCompile(`^{{\s*\.Env\.[^.\s}]+\s*}}$`)

func config(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../.goreleaser.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func value(line string) string {
	v := strings.TrimSpace(line[strings.Index(line, ":")+1:])
	return strings.Trim(v, `'"`)
}

func TestTokensAreASingleEnvReference(t *testing.T) {
	found := 0
	for _, line := range strings.Split(config(t), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "token:") {
			found++
			if v := value(line); !envOnlyRe.MatchString(v) {
				t.Errorf("token %q is not a single {{ .Env.NAME }} reference; GoReleaser rejects it at publish time", v)
			}
		}
	}
	if found == 0 {
		t.Fatal("no token field found; did the Homebrew cask section move?")
	}
}

// Only functions GoReleaser's template engine provides may appear; envOrDefault
// (used in v0.1.0) is not one of them.
func TestNoUnknownTemplateFunctions(t *testing.T) {
	for _, bad := range []string{"envOrDefault", "env \""} {
		if strings.Contains(config(t), bad) {
			t.Errorf(".goreleaser.yaml uses %q, which GoReleaser does not define", bad)
		}
	}
}
