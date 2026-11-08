package cli

import (
	"fmt"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/store"
	"github.com/spf13/cobra"
)

func (a *app) regroupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "regroup NAME GROUP",
		Short: "Move an alias to another group",
		Long: `Move alias NAME to GROUP. Groups are lowercase letters, digits, - and _. The
choice is kept when the command is updated later.`,
		Example: `  ali regroup hgp k8s`,
		Args:    args(2),
		RunE: func(_ *cobra.Command, argv []string) error {
			name, grp := argv[0], argv[1]
			if err := alias.ValidateGroup(grp); err != nil {
				return err
			}
			return a.st.Update(func(d *store.Data) error {
				cur, ok := d.Get(name)
				if !ok {
					return fmt.Errorf("no alias %q", name)
				}
				cur.Group = grp
				d.Set(cur, true)
				fmt.Fprintf(a.out, "moved %s to %s\n", name, grp)
				return nil
			})
		},
	}
}
