package cli

import (
	"fmt"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/spf13/cobra"
)

func (a *app) lsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "ls [GROUP]",
		Aliases: []string{"list"},
		Short:   "List aliases, grouped (optionally only one group)",
		Long: `List all aliases under group headings, sorted by group and name, with commands
aligned. Give a GROUP to list only that group. Control characters in commands are
shown as escapes such as \x1b. Use --json for machine-readable output.`,
		Example: `  ali ls
  ali ls helm
  ali ls --json | jq -r '.[].name'`,
		Args: maxArgs(1),
		RunE: func(cmd *cobra.Command, argv []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			list := d.All()
			if len(argv) == 1 {
				var only []alias.Alias
				for _, x := range list {
					if x.Group == argv[0] {
						only = append(only, x)
					}
				}
				if len(only) == 0 {
					return fmt.Errorf("no group %q (see: ali groups)", argv[0])
				}
				list = only
			}
			if asJSON {
				return a.printJSON(nonNil(list))
			}
			if len(list) == 0 {
				fmt.Fprintln(a.out, "no aliases yet. try: ali add gp 'git pull'")
				return nil
			}
			a.printGrouped(list)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}
