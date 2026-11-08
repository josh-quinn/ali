package shadow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShadows(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "mytool")
	os.WriteFile(exe, []byte("#!/bin/sh\n"), 0o755)
	os.WriteFile(filepath.Join(dir, "notexec"), []byte("x"), 0o644)
	t.Setenv("PATH", dir)

	if what := Shadows("mytool"); what == "" {
		t.Error("executable on PATH not detected")
	}
	if what := Shadows("notexec"); what != "" {
		t.Errorf("non-executable flagged: %s", what)
	}
	if what := Shadows("cd"); what == "" {
		t.Error("builtin cd not detected")
	}
	if what := Shadows("if"); what == "" {
		t.Error("reserved word not detected")
	}
	if what := Shadows("totally-free-name"); what != "" {
		t.Errorf("free name flagged: %s", what)
	}
}
