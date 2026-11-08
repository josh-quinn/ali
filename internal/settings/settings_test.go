package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsWhenMissing(t *testing.T) {
	s, err := Load(t.TempDir())
	if err != nil || s.Expand != Inline || s.Prompt != PromptRight {
		t.Errorf("%+v %v", s, err)
	}
}

func TestSaveLoadAndPermissions(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, Settings{Expand: Runtime, Prompt: PromptOff}); err != nil {
		t.Fatal(err)
	}
	s, err := Load(dir)
	if err != nil || s.Expand != Runtime || s.Prompt != PromptOff {
		t.Errorf("%+v %v", s, err)
	}
	if st, _ := os.Stat(filepath.Join(dir, FileName)); st.Mode().Perm() != 0o600 {
		t.Errorf("mode %v", st.Mode().Perm())
	}
}

func TestInvalidValues(t *testing.T) {
	for _, body := range []string{"expand = \"sometimes\"\n", "prompt = \"left\"\n", "expand = 1\n", "[x\n", "unknown = \"y\"\n"} {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, FileName), []byte(body), 0o600)
		if _, err := Load(dir); err == nil {
			t.Errorf("%q: want error", body)
		}
	}
	if err := (Settings{Expand: "x", Prompt: PromptRight}).Validate(); err == nil {
		t.Error("bad expand accepted")
	}
}

func TestSetByKey(t *testing.T) {
	s := Defaults()
	if err := s.SetKey("expand", "runtime"); err != nil || s.Expand != Runtime {
		t.Errorf("%+v %v", s, err)
	}
	if err := s.SetKey("prompt", "nope"); err == nil {
		t.Error("bad prompt accepted")
	}
	if err := s.SetKey("colour", "red"); err == nil {
		t.Error("unknown key accepted")
	}
	if v, err := s.GetKey("expand"); err != nil || v != "runtime" {
		t.Errorf("%q %v", v, err)
	}
}

func TestGlobalEnv(t *testing.T) {
	dir := t.TempDir()
	if g, err := GlobalEnv(dir); err != nil || g != "" {
		t.Errorf("missing file: %q %v", g, err)
	}
	if err := SetGlobalEnv(dir, "prod"); err != nil {
		t.Fatal(err)
	}
	if g, _ := GlobalEnv(dir); g != "prod" {
		t.Errorf("got %q", g)
	}
	if err := SetGlobalEnv(dir, ""); err != nil {
		t.Fatal(err)
	}
	if g, _ := GlobalEnv(dir); g != "" {
		t.Errorf("not cleared: %q", g)
	}
	if err := SetGlobalEnv(dir, "Bad Name"); err == nil {
		t.Error("invalid name accepted")
	}
	os.WriteFile(filepath.Join(dir, GlobalEnvFile), []byte("$(rm -rf ~)\n"), 0o600)
	if _, err := GlobalEnv(dir); err == nil {
		t.Error("hand-edited invalid global env accepted")
	}
	if st, _ := os.Stat(filepath.Join(dir, GlobalEnvFile)); st.Mode().Perm() != 0o600 {
		t.Errorf("mode %v", st.Mode().Perm())
	}
}
