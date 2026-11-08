package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/cli"
	"github.com/josh-quinn/ali/internal/store"
)

func TestGeneratesAPageForEveryCommandAndTopic(t *testing.T) {
	dir := t.TempDir()
	if err := generate(dir, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	root := cli.NewRoot("1.2.3", store.New(t.TempDir()), io.Discard, io.Discard)
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Hidden || c.Name() == "help" {
			return
		}
		name := strings.ReplaceAll(c.CommandPath(), " ", "-") + ".1"
		if c.IsAdditionalHelpTopicCommand() {
			name = strings.TrimSuffix(name, ".1") + ".7" // topics are section 7
		}
		page, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("no man page for %q: %v", c.CommandPath(), err)
		} else {
			if c.Example != "" && !strings.Contains(string(page), "EXAMPLE") {
				t.Errorf("%s has no EXAMPLE section", name)
			}
			// every argument placeholder in Use must survive into the page
			for _, w := range strings.Fields(c.Use)[1:] {
				w = strings.Trim(w, "[].")
				if w != "" && !strings.HasPrefix(w, "-") && !strings.Contains(string(page), w) {
					t.Errorf("%s synopsis lost %q (Use: %q)", name, w, c.Use)
				}
			}
		}
		for _, s := range c.Commands() {
			walk(s)
		}
	}
	walk(root)

	main, _ := os.ReadFile(filepath.Join(dir, "ali.1"))
	if !strings.Contains(string(main), "1.2.3") {
		t.Error("ali.1 does not carry the version")
	}
}

func TestReproducibleDateFromSourceDateEpoch(t *testing.T) {
	t.Setenv("SOURCE_DATE_EPOCH", "1767225600") // 2026-01-01
	dir := t.TempDir()
	if err := generate(dir, "x"); err != nil {
		t.Fatal(err)
	}
	page, _ := os.ReadFile(filepath.Join(dir, "ali.1"))
	if !strings.Contains(string(page), "Jan 2026") {
		t.Errorf("date not taken from SOURCE_DATE_EPOCH:\n%.300s", page)
	}
}
