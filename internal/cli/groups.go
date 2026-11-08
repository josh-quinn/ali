package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func (a *app) groupsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "List groups with alias counts",
		Long: `List every group with the number of aliases in it. Use --json for
machine-readable output.`,
		Example: `  ali groups
  ali groups --json`,
		Args: args(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			counts := map[string]int{}
			for _, x := range d.All() {
				counts[x.Group]++
			}
			names := make([]string, 0, len(counts))
			width := 0
			for g := range counts {
				names = append(names, g)
				if len(g) > width {
					width = len(g)
				}
			}
			sort.Strings(names)
			if asJSON {
				type row struct {
					Group string `json:"group"`
					Count int    `json:"count"`
				}
				rows := []row{}
				for _, g := range names {
					rows = append(rows, row{g, counts[g]})
				}
				return a.printJSON(rows)
			}
			for _, g := range names {
				fmt.Fprintf(a.out, "%-*s  %d\n", width, g, counts[g])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}
