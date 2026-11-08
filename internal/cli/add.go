package cli

import (
	"fmt"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/group"
	"github.com/josh-quinn/ali/internal/shadow"
	"github.com/josh-quinn/ali/internal/store"
	"github.com/spf13/cobra"
)

func (a *app) addCmd() *cobra.Command {
	var grp string
	var force bool
	cmd := &cobra.Command{
		Use:   "add <name> <command> [-g group] [--force]",
		Short: "Create an alias (group defaults to the command's first word)",
		Args:  args(2),
		RunE: func(cmd *cobra.Command, argv []string) error {
			name, command := argv[0], argv[1]
			if err := alias.ValidateName(name); err != nil {
				return err
			}
			if err := alias.ValidateCommand(command); err != nil {
				return err
			}
			pinned := grp != ""
			if pinned {
				if err := alias.ValidateGroup(grp); err != nil {
					return err
				}
			} else {
				grp = group.Infer(command)
			}
			return a.st.Update(func(d *store.Data) error {
				prev, exists := d.Get(name)
				if exists && !force {
					return fmt.Errorf("alias %s already exists in group %s (%s); use `ali update %s <command>` or --force", name, prev.Group, display(prev.Command), name)
				}
				if !exists {
					if what := shadow.Shadows(name); what != "" {
						fmt.Fprintf(a.err, "warning: %s shadows %s\n", name, what)
					}
				}
				d.Set(alias.Alias{Name: name, Command: command, Group: grp}, pinned)
				fmt.Fprintf(a.out, "added %s (%s): %s\n", name, grp, display(command))
				return nil
			})
		},
	}
	cmd.Flags().StringVarP(&grp, "group", "g", "", "group name (default: first word of the command)")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing alias")
	return cmd
}
