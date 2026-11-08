package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImport(t *testing.T) {
	e := newEnv(t)
	rc := filepath.Join(e.dir, ".zshrc")
	orig := "alias k=kubectl\nalias gs='git status'\nalias gp='git pull'\nalias bad=\"echo $HOME\"\nexport X=1\n"
	os.WriteFile(rc, []byte(orig), 0o644)
	e.must("add", "gp", "git pull --rebase")

	out := e.must("import", "--dry-run")
	if !strings.Contains(out, "would import 2") || !strings.Contains(out, "skipped 2") {
		t.Errorf("dry-run summary: %q", out)
	}
	if d, _ := e.st.Load(); len(d.All()) != 1 {
		t.Error("dry-run changed store")
	}

	out = e.must("import")
	if !strings.Contains(out, "imported 2") {
		t.Errorf("summary: %q", out)
	}
	d, _ := e.st.Load()
	if a, ok := d.Get("k"); !ok || a.Group != "kubectl" {
		t.Errorf("k: %+v", a)
	}
	if a, _ := d.Get("gp"); a.Command != "git pull --rebase" {
		t.Error("conflict overwrote existing alias")
	}
	if !strings.Contains(out, "alias k=kubectl") {
		t.Errorf("originals not listed: %q", out)
	}
	if b, _ := os.ReadFile(rc); string(b) != orig {
		t.Error("import modified the source file")
	}
	matches, _ := filepath.Glob(rc + ".ali-backup-*")
	if len(matches) != 1 {
		t.Errorf("want 1 backup, got %v", matches)
	}
}

func TestImportReportSanitisesSkippedLines(t *testing.T) {
	e := newEnv(t)
	rc := filepath.Join(e.dir, ".zshrc")
	os.WriteFile(rc, []byte("alias bad=\"echo $HOME \x1b[31m\"\n"), 0o644)
	out := e.must("import", "--dry-run")
	if strings.ContainsRune(out, '\x1b') {
		t.Errorf("import report leaked an escape: %q", out)
	}
}
