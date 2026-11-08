package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/doctor"
	"github.com/josh-quinn/ali/internal/shellconf"
)

func (a *app) doctorCmd() *cobra.Command {
	var state string
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check that ali is set up correctly in your shell",
		Long: `Check your setup and say how to fix anything wrong:

  block present in .zshrc               "ali init zsh" has been run
  block up to date                      the block matches this version of ali
  block loads after Oh My Zsh           otherwise Oh My Zsh replaces ali's aliases
  aliases.toml readable                 your alias file is valid
  aliases not redefined after the block nothing later in .zshrc changes ali's aliases

The last check needs to know what your shell loaded, so it only runs when you type
"ali doctor" in your shell (the ali shell function passes that information along);
from a script it is reported as skipped. Exits with status 1 if any check fails.`,
		Example: `  ali doctor`,
		Args:    args(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			in := doctor.Input{}
			if state == "-" {
				raw, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), 16<<20))
				if err != nil {
					return err
				}
				in.Overridden = doctor.ParseState(string(raw))
			}
			if rc, err := shellconf.Path(); err != nil {
				in.RCMissing = true
			} else if b, err := os.ReadFile(rc); err != nil {
				in.RCMissing = true
			} else {
				in.RC = string(b)
			}
			_, in.StoreErr = a.st.Load()

			checks := doctor.Run(in)
			failed := 0
			for _, c := range checks {
				line := fmt.Sprintf("%-7s %s", c.Status, c.Name)
				if c.Detail != "" {
					line += ": " + display(c.Detail)
				}
				fmt.Fprintln(a.out, line)
				if c.Status == doctor.Fail {
					failed++
					fmt.Fprintf(a.out, "        fix: %s\n", c.Fix)
				}
			}
			if failed > 0 {
				return fmt.Errorf("%d check(s) failed", failed)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&state, "shell-state", "", "internal: read shell state from stdin (used by the ali shell function)")
	cmd.Flags().MarkHidden("shell-state")
	return cmd
}
