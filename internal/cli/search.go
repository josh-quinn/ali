package cli

import (
	"fmt"
	"strings"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/spf13/cobra"
)

func (a *app) searchCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "search <text>",
		Short: "Find aliases whose name, command or group contains text",
		Long: `Search alias names, commands and groups for TEXT, ignoring case. Results are shown
like "ali ls". Exits with status 1 when nothing matches, so it works in scripts.`,
		Example: `  ali search pods
  ali search GIT --json`,
		Args: args(1),
		RunE: func(cmd *cobra.Command, argv []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			needle := strings.ToLower(argv[0])
			var hits []alias.Alias
			for _, x := range d.All() {
				if strings.Contains(strings.ToLower(x.Name+"\x00"+x.Command+"\x00"+x.Group), needle) {
					hits = append(hits, x)
				}
			}
			if len(hits) == 0 {
				return fmt.Errorf("no aliases match %q", argv[0])
			}
			if asJSON {
				return a.printJSON(hits)
			}
			a.printGrouped(hits)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}
