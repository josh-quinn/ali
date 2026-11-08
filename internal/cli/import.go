package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/group"
	"github.com/josh-quinn/ali/internal/importer"
	"github.com/josh-quinn/ali/internal/shellconf"
	"github.com/josh-quinn/ali/internal/store"
	"github.com/spf13/cobra"
)

func (a *app) importCmd() *cobra.Command {
	var file string
	var dry bool
	cmd := &cobra.Command{
		Use:   "import [--file PATH] [--dry-run]",
		Short: "Copy existing `alias` lines from your .zshrc into ali (originals are left alone)",
		Long: `Read simple "alias name=value" lines from your .zshrc (or --file) and add them to
ali, grouped automatically. Your file is never modified: ali prints the lines you
can delete afterwards. Names already in ali are skipped. Lines ali cannot read with
certainty (multi-line values, $ or backticks inside double quotes, alias -g, several
definitions on one line, unquoted ; | & < > ( )) are reported, never guessed.`,
		Example: `  ali import --dry-run
  ali import
  ali import --file ~/.oh-my-zsh/custom/aliases.zsh`,
		Args: args(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if file == "" {
				p, err := shellconf.Path()
				if err != nil {
					return err
				}
				file = p
			}
			src, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			parsed := importer.Parse(string(src))
			var imported []importer.Entry
			skipped := append([]importer.Skip(nil), parsed.Skipped...)

			plan := func(d *store.Data) {
				for _, e := range parsed.Entries {
					if err := alias.ValidateName(e.Name); err != nil {
						skipped = append(skipped, importer.Skip{Line: e.Line, LineNo: e.LineNo, Reason: err.Error()})
						continue
					}
					if _, exists := d.Get(e.Name); exists {
						skipped = append(skipped, importer.Skip{Line: e.Line, LineNo: e.LineNo, Reason: "already in ali"})
						continue
					}
					d.Set(alias.Alias{Name: e.Name, Command: e.Command, Group: group.Infer(e.Command)}, false)
					imported = append(imported, e)
				}
			}
			if dry {
				d, err := a.st.Load()
				if err != nil {
					return err
				}
				plan(d)
			} else if err := a.st.Update(func(d *store.Data) error { plan(d); return nil }); err != nil {
				return err
			}
			if !dry && len(imported) > 0 {
				bk, err := shellconf.Backup(file)
				if err != nil {
					return err
				}
				fmt.Fprintf(a.err, "backup of %s: %s\n", file, bk)
			}

			sort.Slice(skipped, func(i, j int) bool { return skipped[i].LineNo < skipped[j].LineNo })
			verb := "imported"
			if dry {
				verb = "would import"
			}
			fmt.Fprintf(a.out, "%s %d, skipped %d\n", verb, len(imported), len(skipped))
			for _, s := range skipped {
				fmt.Fprintf(a.out, "  skipped line %d (%s): %s\n", s.LineNo, s.Reason, display(s.Line))
			}
			if len(imported) > 0 {
				fmt.Fprintf(a.out, "\nthe originals are still in %s; delete these lines to avoid duplicates:\n", file)
				for _, e := range imported {
					fmt.Fprintf(a.out, "  line %d: %s\n", e.LineNo, display(e.Line))
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "file to read (default: your .zshrc)")
	cmd.Flags().BoolVar(&dry, "dry-run", false, "show what would be imported without writing")
	return cmd
}
