package cli

import (
	"fmt"

	"github.com/josh-quinn/ali/internal/render"
	"github.com/josh-quinn/ali/internal/shellconf"
	"github.com/josh-quinn/ali/internal/store"
	"github.com/spf13/cobra"
)

func (a *app) initCmd() *cobra.Command {
	var dry bool
	cmd := &cobra.Command{
		Use:   "init zsh [--dry-run]",
		Short: "One-time setup: add the ali block to your .zshrc (with backup)",
		Args:  args(1),
		RunE: func(cmd *cobra.Command, argv []string) error {
			if argv[0] != "zsh" {
				return &UsageError{fmt.Errorf("unsupported shell %q (only zsh for now)", argv[0])}
			}
			rc := shellconf.Path()
			if dry {
				fmt.Fprintf(a.out, "would append to %s:\n\n%s", rc, render.Block())
				return nil
			}
			// Make sure the file the block sources exists (and the store is valid).
			if err := a.st.Update(func(*store.Data) error { return nil }); err != nil {
				return err
			}
			res, err := shellconf.Install(rc)
			if err != nil {
				return err
			}
			if !res.Changed {
				fmt.Fprintf(a.out, "already set up (%s)\n", rc)
				return nil
			}
			fmt.Fprintf(a.out, "added the ali block to %s\n", rc)
			if res.Backup != "" {
				fmt.Fprintf(a.out, "backup: %s\n", res.Backup)
			}
			fmt.Fprintf(a.out, "open a new terminal or run: source %s\n", rc)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dry, "dry-run", false, "show what would change without writing")
	return cmd
}
