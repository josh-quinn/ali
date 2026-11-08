package cli

import (
	"fmt"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/group"
	"github.com/josh-quinn/ali/internal/store"
	"github.com/spf13/cobra"
)

func (a *app) updateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update NAME COMMAND",
		Short: "Change what an alias runs",
		Long: `Replace the command of an existing alias. The group is re-inferred from the new
command unless you set it explicitly with "ali add -g" or "ali regroup".`,
		Example: `  ali update gp 'git pull --rebase'`,
		Args:    args(2),
		RunE: func(cmd *cobra.Command, argv []string) error {
			name, command := argv[0], argv[1]
			if err := alias.ValidateCommand(command); err != nil {
				return err
			}
			return a.st.Update(func(d *store.Data) error {
				cur, ok := d.Get(name)
				if !ok {
					return fmt.Errorf("no alias %q", name)
				}
				pinned := d.Pinned(name)
				if !pinned {
					cur.Group = group.Infer(command)
				}
				cur.Command = command
				d.Set(cur, pinned)
				fmt.Fprintf(a.out, "updated %s (%s): %s\n", name, cur.Group, display(command))
				return nil
			})
		},
	}
}
