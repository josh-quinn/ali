package cli

import (
	"io"
	"testing"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/store"
)

// Constitution v1.2.0 / ADR 0038: every command documents itself fully.
func TestEveryCommandHasSummaryDescriptionAndExample(t *testing.T) {
	root := NewRoot("test", store.New(t.TempDir()), io.Discard, io.Discard)
	root.InitDefaultHelpCmd()
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if c.Name() == "help" && c.Parent() == root { // cobra's built-in
			return
		}
		if c.Short == "" {
			t.Errorf("%q has no Short", c.CommandPath())
		}
		if c.Long == "" {
			t.Errorf("%q has no Long", c.CommandPath())
		}
		if c.Example == "" && c.Runnable() {
			t.Errorf("%q has no Example", c.CommandPath())
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
}
