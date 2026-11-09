// Package release checks .goreleaser.yaml for mistakes that `goreleaser check`
// and snapshot builds cannot catch, because the fields are only evaluated when
// publishing. v0.1.0 and v0.1.1 shipped without their Homebrew cask because of
// two such mistakes.
package release

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

// installSteps returns the cask's install_steps block as GoReleaser renders it:
// {{ "x" }} escapes become x.
func installSteps(t *testing.T) string {
	t.Helper()
	lines := strings.Split(config(t), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "install_steps: |" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		var block []string
		for _, l := range lines[i+1:] {
			if strings.TrimSpace(l) != "" && len(l)-len(strings.TrimLeft(l, " ")) <= indent {
				break
			}
			block = append(block, l)
		}
		return regexp.MustCompile(`{{ "([^"]*)" }}`).ReplaceAllString(strings.Join(block, "\n"), "$1")
	}
	t.Fatal("no install_steps found; did the Homebrew cask section move?")
	return ""
}

// v0.1.2's cask used Ruby (#{staged_path}, system_command) inside
// postflight_steps, which Homebrew's declarative steps DSL rejects: the cask
// was unreadable and could not be installed.
func TestCaskInstallStepsAreDeclarative(t *testing.T) {
	steps := installSteps(t)
	for _, bad := range []string{"#{", "system_command", "OS.mac?"} {
		if strings.Contains(steps, bad) {
			t.Errorf("install_steps uses %q; Homebrew's install steps DSL only accepts declarative steps", bad)
		}
	}
	brew, err := exec.LookPath("brew")
	if err != nil {
		t.Skip("brew not installed; cannot load the steps in Homebrew")
	}
	cask := "cask \"ali\" do\n  version \"0.0.0\"\n  sha256 :no_check\n  url \"https://example.com/ali.tar.gz\"\n" +
		"  name \"ali\"\n  desc \"test\"\n  homepage \"https://example.com\"\n  binary \"ali\"\n" +
		"  postflight_steps do\n" + steps + "\n  end\nend\n"
	f := filepath.Join(t.TempDir(), "ali.rb")
	if err := os.WriteFile(f, []byte(cask), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(brew, "ruby", "-e",
		`exit 77 unless defined?(Cask::Artifact::PostflightSteps)
		Cask::CaskLoader::FromContentLoader.new(File.read(ARGV[0])).load(config: nil)`, f).CombinedOutput()
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 77 {
		t.Skip("this Homebrew predates install steps")
	}
	if err != nil {
		t.Fatalf("Homebrew cannot read the cask: %v\n%s", err, out)
	}
}
