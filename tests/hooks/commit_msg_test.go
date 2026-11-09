package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runCommitMsg runs .githooks/commit-msg on a message file, as git does.
func runCommitMsg(t *testing.T, msg string) (string, error) {
	t.Helper()
	hook, _ := filepath.Abs("../../.githooks/commit-msg")
	f := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(f, []byte(msg), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(hook, f).CombinedOutput()
	return string(out), err
}

func TestCommitMsgAcceptsConventionalCommits(t *testing.T) {
	for _, msg := range []string{
		"feat(cli): add ali env\n",
		"fix: handle empty store\n\nLonger body.\n",
		"docs(adr): record the decision",
		"build(deps): bump golang.org/x/sys from 0.1.0 to 0.2.0\n",
		"refactor!: drop store format v1\n",
		"chore(github)!: require reviews\n",
		"revert: feat(cli): add ali env\n",
		"# a git comment line first\nci: run fuzz targets\n",
		"Merge branch 'main' into feature\n",
		"Merge pull request #12 from someone/branch\n",
		"fixup! feat(cli): add ali env\n",
		"squash! fix: handle empty store\n",
	} {
		if out, err := runCommitMsg(t, msg); err != nil {
			t.Errorf("rejected %q: %v\n%s", msg, err, out)
		}
	}
}

func TestCommitMsgRejectsOtherMessages(t *testing.T) {
	for _, msg := range []string{
		"add a feature\n",
		"Feat: capitalised type\n",
		"feat:missing space\n",
		"feat(): empty scope\n",
		"feat(cli) missing colon\n",
		"wip(cli): unknown type\n",
		"feat: \n",
		"\n\n",
		"# only a comment\n",
	} {
		if _, err := runCommitMsg(t, msg); err == nil {
			t.Errorf("accepted %q", msg)
		}
	}
}

func TestCommitMsgExplainsTheFormat(t *testing.T) {
	out, err := runCommitMsg(t, "add a feature\n")
	if err == nil {
		t.Fatal("accepted a non-conventional message")
	}
	for _, want := range []string{"add a feature", "type(scope): subject", "CONTRIBUTING.md"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q does not mention %q", out, want)
		}
	}
}
