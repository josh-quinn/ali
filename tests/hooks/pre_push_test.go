// Package hooks tests the repository's git hooks in throwaway repositories.
package hooks

import (
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type repo struct {
	t      *testing.T
	dir    string
	remote string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	if _, err := exec.LookPath("gitleaks"); err != nil {
		t.Skip("gitleaks not installed")
	}
	hook, _ := filepath.Abs("../../.githooks/pre-push")
	r := &repo{t: t, dir: t.TempDir(), remote: t.TempDir()}
	r.git(r.remote, "init", "-q", "--bare")
	r.git(r.dir, "init", "-q", "-b", "main")
	r.git(r.dir, "config", "user.email", "test@example.com")
	r.git(r.dir, "config", "user.name", "test")
	r.git(r.dir, "config", "commit.gpgsign", "false")
	r.git(r.dir, "remote", "add", "origin", r.remote)
	os.MkdirAll(filepath.Join(r.dir, ".git", "hooks"), 0o755)
	b, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(r.dir, ".git", "hooks", "pre-push"), b, 0o755)
	return r
}

func (r *repo) git(dir string, args ...string) (string, error) {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (r *repo) commit(name, body string) {
	r.t.Helper()
	os.WriteFile(filepath.Join(r.dir, name), []byte(body), 0o644)
	r.git(r.dir, "add", name)
	if out, err := r.git(r.dir, "commit", "-q", "-m", "add "+name); err != nil {
		r.t.Fatalf("commit: %v %s", err, out)
	}
}

func (r *repo) remoteHas(ref string) bool {
	_, err := r.git(r.remote, "rev-parse", "--verify", "-q", ref)
	return err == nil
}

// fakeToken looks like a GitHub personal access token. It is generated at run
// time (fixed seed, so reproducible) so this repository's own source never
// contains anything token-like; a repetitive fake would be ignored by gitleaks
// as too low-entropy.
func fakeToken() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rng := rand.New(rand.NewPCG(1, 2))
	b := make([]byte, 36)
	for i := range b {
		b[i] = alphabet[rng.IntN(len(alphabet))]
	}
	return "gh" + "p_" + string(b)
}

func TestCleanPushSucceeds(t *testing.T) {
	r := newRepo(t)
	r.commit("README", "hello\n")
	if out, err := r.git(r.dir, "push", "-q", "origin", "main"); err != nil {
		t.Fatalf("clean push blocked: %v\n%s", err, out)
	}
	if !r.remoteHas("main") {
		t.Error("remote did not receive main")
	}
}

func TestPushWithASecretIsBlocked(t *testing.T) {
	r := newRepo(t)
	r.commit("README", "hello\n")
	r.git(r.dir, "push", "-q", "origin", "main")
	before, _ := r.git(r.remote, "rev-parse", "main")
	r.commit("config.env", "GITHUB_TOKEN="+fakeToken()+"\n")
	out, err := r.git(r.dir, "push", "-q", "origin", "main")
	if err == nil {
		t.Fatalf("push with a token was not blocked:\n%s", out)
	}
	if !strings.Contains(out, "secret") {
		t.Errorf("no explanation for the block:\n%s", out)
	}
	if strings.Contains(out, fakeToken()) {
		t.Error("the hook printed the secret itself; findings must be redacted")
	}
	if after, _ := r.git(r.remote, "rev-parse", "main"); after != before {
		t.Error("remote changed despite the block")
	}
}

// A secret added and removed again before pushing still travels in history.
func TestSecretInAnEarlierUnpushedCommitIsBlocked(t *testing.T) {
	r := newRepo(t)
	r.commit("config.env", "TOKEN="+fakeToken()+"\n")
	r.commit("config.env", "TOKEN=\n")
	if out, err := r.git(r.dir, "push", "-q", "origin", "main"); err == nil {
		t.Fatalf("history with a removed token was pushed:\n%s", out)
	}
}

func TestNewBranchIsScanned(t *testing.T) {
	r := newRepo(t)
	r.commit("README", "hello\n")
	r.git(r.dir, "push", "-q", "origin", "main")
	r.git(r.dir, "switch", "-q", "-c", "feature")
	r.commit("key.txt", "token: "+fakeToken()+"\n")
	if out, err := r.git(r.dir, "push", "-q", "origin", "feature"); err == nil || r.remoteHas("feature") {
		t.Fatalf("new branch with a token was pushed:\n%s", out)
	}
}

func TestDeletingABranchIsAllowed(t *testing.T) {
	r := newRepo(t)
	r.commit("README", "hello\n")
	r.git(r.dir, "push", "-q", "origin", "main", "main:old")
	if out, err := r.git(r.dir, "push", "-q", "origin", "--delete", "old"); err != nil {
		t.Fatalf("branch deletion blocked: %v\n%s", err, out)
	}
}
