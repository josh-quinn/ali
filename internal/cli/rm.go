package cli

import (
	"fmt"
	"strings"

	"github.com/josh-quinn/ali/internal/store"
	"github.com/spf13/cobra"
)

func (a *app) rmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm <name>...",
		Aliases: []string{"remove", "delete"},
		Short:   "Delete aliases",
		Args:    minArgs(1),
		RunE: func(cmd *cobra.Command, argv []string) error {
			var missing []string
			err := a.st.Update(func(d *store.Data) error {
				for _, name := range argv {
					cur, ok := d.Get(name)
					if !ok {
						missing = append(missing, name)
						continue
					}
					d.Delete(name)
					fmt.Fprintf(a.out, "removed %s: %s\n", name, display(cur.Command))
				}
				return nil
			})
			if err != nil {
				return err
			}
			if len(missing) > 0 {
				return fmt.Errorf("no such alias: %s", strings.Join(missing, ", "))
			}
			return nil
		},
	}
}
