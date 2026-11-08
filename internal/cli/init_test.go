package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitDryRunAndInstall(t *testing.T) {
	e := newEnv(t)
	rc := filepath.Join(e.dir, ".zshrc")
	os.WriteFile(rc, []byte("export A=1\n"), 0o644)

	out := e.must("init", "zsh", "--dry-run")
	if !strings.Contains(out, "# >>> ali >>>") {
		t.Errorf("dry run output: %q", out)
	}
	if b, _ := os.ReadFile(rc); string(b) != "export A=1\n" {
		t.Error("dry-run modified .zshrc")
	}

	out = e.must("init", "zsh")
	if !strings.Contains(out, "backup") {
		t.Errorf("install output: %q", out)
	}
	if b, _ := os.ReadFile(rc); !strings.Contains(string(b), "# >>> ali >>>") {
		t.Error("block not installed")
	}
	if out := e.must("init", "zsh"); !strings.Contains(out, "already set up") {
		t.Errorf("second run: %q", out)
	}
	if _, err := os.Stat(e.st.ZshFile()); err != nil {
		t.Error("aliases.zsh should exist after init so the block can source it")
	}
	if _, _, err := e.run("init", "fish"); err == nil {
		t.Error("only zsh is supported")
	}
}

func TestInitReportsUpgradeAndRefusesDamagedBlock(t *testing.T) {
	e := newEnv(t)
	rc := filepath.Join(e.dir, ".zshrc")
	os.WriteFile(rc, []byte("export A=1\n# >>> ali >>>\nold\n# <<< ali <<<\n"), 0o644)
	out := e.must("init", "zsh")
	if !strings.Contains(out, "updated the ali block in "+rc) || !strings.Contains(out, "backup:") {
		t.Errorf("upgrade output: %q", out)
	}

	os.WriteFile(rc, []byte("# >>> ali >>>\nno end marker\n"), 0o644)
	_, _, err := e.run("init", "zsh")
	if err == nil || !strings.Contains(err.Error(), "markers") {
		t.Errorf("want damaged-block error, got %v", err)
	}
}
