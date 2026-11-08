package store

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/josh-quinn/ali/internal/alias"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return New(t.TempDir())
}

func TestMissingFileIsEmpty(t *testing.T) {
	s := newStore(t)
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.All()) != 0 {
		t.Errorf("want empty, got %v", d.All())
	}
}

func TestRoundTrip(t *testing.T) {
	s := newStore(t)
	err := s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "git"}, false)
		d.Set(alias.Alias{Name: "hgp", Command: "helm get pods", Group: "k8s"}, true)
		d.Set(alias.Alias{Name: "q", Command: "echo 'it'\\''s' \"$X\"\nnext", Group: "echo"}, false)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	a, ok := d.Get("q")
	if !ok || a.Command != "echo 'it'\\''s' \"$X\"\nnext" {
		t.Errorf("special-char command lost: %+v", a)
	}
	if !d.Pinned("hgp") || d.Pinned("gp") {
		t.Error("pinned flags wrong")
	}
	if g, _ := d.Get("hgp"); g.Group != "k8s" {
		t.Errorf("group = %q", g.Group)
	}
	z, err := os.ReadFile(filepath.Join(s.Dir, "aliases.zsh"))
	if err != nil || !strings.Contains(string(z), "alias -- gp='git pull'") {
		t.Errorf("aliases.zsh not rendered: %v %s", err, z)
	}
}

func TestDeleteDropsEmptyGroup(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "git"}, false)
		return nil
	})
	s.Update(func(d *Data) error { d.Delete("gp"); return nil })
	raw, _ := os.ReadFile(s.File())
	if strings.Contains(string(raw), "[git]") {
		t.Errorf("empty group table kept:\n%s", raw)
	}
}

func TestMoveBetweenGroupsKeepsSingleCopy(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "git"}, false)
		return nil
	})
	s.Update(func(d *Data) error {
		d.Set(alias.Alias{Name: "gp", Command: "git pull", Group: "tools"}, true)
		return nil
	})
	d, _ := s.Load()
	if len(d.All()) != 1 {
		t.Errorf("want 1 alias, got %v", d.All())
	}
}

func TestHandEditedFileIsHonoured(t *testing.T) {
	s := newStore(t)
	os.WriteFile(s.File(), []byte("version = 1\n\n[git]\ngs = \"git status\"\n"), 0o600)
	d, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if a, ok := d.Get("gs"); !ok || a.Group != "git" {
		t.Errorf("hand-edited alias not loaded: %+v", a)
	}
}

func TestInvalidFilesError(t *testing.T) {
	cases := map[string]string{
		"syntax":    "version = 1\n[git\n",
		"duplicate": "version = 1\n[a]\nx = \"1\"\n[b]\nx = \"2\"\n",
		"badname":   "version = 1\n[a]\n\"bad name\" = \"1\"\n",
		"nonstring": "version = 1\n[a]\nx = 3\n",
		"version":   "version = 99\n",
		"toplevel":  "version = 1\nstray = \"x\"\n",
	}
	for name, body := range cases {
		s := newStore(t)
		os.WriteFile(s.File(), []byte(body), 0o600)
		if _, err := s.Load(); err == nil {
			t.Errorf("%s: want error", name)
		} else if !strings.Contains(err.Error(), "aliases.toml") {
			t.Errorf("%s: error should name the file: %v", name, err)
		}
		// a failed update must not overwrite the broken file
		before, _ := os.ReadFile(s.File())
		s.Update(func(d *Data) error { return nil })
		after, _ := os.ReadFile(s.File())
		if string(before) != string(after) {
			t.Errorf("%s: invalid store was overwritten", name)
		}
	}
}

func TestUpdateErrorLeavesStoreUntouched(t *testing.T) {
	s := newStore(t)
	s.Update(func(d *Data) error { d.Set(alias.Alias{Name: "a", Command: "x", Group: "g"}, false); return nil })
	before, _ := os.ReadFile(s.File())
	err := s.Update(func(d *Data) error { d.Delete("a"); return os.ErrInvalid })
	if err == nil {
		t.Fatal("want error")
	}
	after, _ := os.ReadFile(s.File())
	if string(before) != string(after) {
		t.Error("store changed despite error")
	}
}

func TestConcurrentUpdatesLoseNothing(t *testing.T) {
	s := newStore(t)
	var wg sync.WaitGroup
	names := []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "b1", "b2", "b3"}
	for _, n := range names {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Update(func(d *Data) error {
				d.Set(alias.Alias{Name: n, Command: "echo " + n, Group: "echo"}, false)
				return nil
			}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	d, _ := s.Load()
	if len(d.All()) != len(names) {
		t.Errorf("lost updates: got %d of %d", len(d.All()), len(names))
	}
}

func TestDirResolution(t *testing.T) {
	t.Setenv("ALI_CONFIG_DIR", "/x/ali")
	if got := DefaultDir(); got != "/x/ali" {
		t.Errorf("ALI_CONFIG_DIR: %s", got)
	}
	t.Setenv("ALI_CONFIG_DIR", "")
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	if got := DefaultDir(); got != "/xdg/ali" {
		t.Errorf("XDG: %s", got)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "/h")
	if got := DefaultDir(); got != "/h/.config/ali" {
		t.Errorf("HOME: %s", got)
	}
}
