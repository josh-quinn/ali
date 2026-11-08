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
		Long: `Create an alias NAME that runs COMMAND. Quote COMMAND so your shell passes it as
one argument; it is stored exactly as typed.

The group is the first word of the command (skipping sudo, env, time, VAR=value and
paths), so "helm get pods" goes to group "helm". Use -g to choose a group yourself;
that choice sticks when you later update the command.

If NAME already exists, add fails unless you pass --force. If NAME hides a command or
builtin, or (when run from your shell) an alias or function you already have, ali
prints a warning and still creates the alias.`,
		Example: `  ali add gp 'git pull'
  ali add hgp 'helm get pods' -g k8s
  ali add gp 'git pull --rebase' --force
  ali add hi "echo 'hello world'"        # inner quotes are kept verbatim`,
		Args: args(2),
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
