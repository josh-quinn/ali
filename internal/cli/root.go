// Package cli wires ali's cobra commands to the store. Commands stay thin:
// parse arguments, call the store or a helper package, print the result.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/store"
)

// UsageError marks a mistake in how ali was invoked (exit status 2).
type UsageError struct{ Err error }

func (e *UsageError) Error() string { return e.Err.Error() }
func (e *UsageError) Unwrap() error { return e.Err }

func asUsage(err error, target **UsageError) bool { return errors.As(err, target) }

// IsUsage reports whether err is a usage error.
func IsUsage(err error) bool { var u *UsageError; return asUsage(err, &u) }

type app struct {
	st       *store.Store
	out, err io.Writer
}

// NewRoot builds the ali command tree writing to out and errw.
func NewRoot(version string, st *store.Store, out, errw io.Writer) *cobra.Command {
	a := &app{st: st, out: out, err: errw}
	root := &cobra.Command{
		Use:   "ali",
		Short: "Create, find and organise shell aliases without touching .zshrc",
		Long: `ali manages your shell aliases from the command line. Aliases are saved in
~/.config/ali/aliases.toml (plain TOML you can edit), grouped by tool, and loaded
into every zsh session by a small block that "ali init zsh" adds to your .zshrc.
After setup, changes made with ali take effect in the current terminal at once.

Run "ali help files" to see which files ali reads and writes.`,
		Example: `  ali init zsh                 # one-time setup
  ali add gp 'git pull'        # create an alias (group: git)
  ali ls                       # list everything, grouped
  ali search pods              # find by name, command or group`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.SetErr(errw)
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return &UsageError{err} })
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(a.addCmd(), a.lsCmd(), a.groupsCmd(), a.searchCmd(),
		a.updateCmd(), a.regroupCmd(), a.rmCmd(), a.initCmd(), a.importCmd(), a.doctorCmd(), a.subCmd(), a.varCmd(), a.expandCmd())
	root.AddCommand(topicCmds()...)
	return root
}

func args(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, got []string) error {
		if len(got) != n {
			return &UsageError{fmt.Errorf("%s needs %d argument(s), got %d\nusage: %s", cmd.Name(), n, len(got), cmd.UseLine())}
		}
		return nil
	}
}

func minArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, got []string) error {
		if len(got) < n {
			return &UsageError{fmt.Errorf("%s needs at least %d argument(s)\nusage: %s", cmd.Name(), n, cmd.UseLine())}
		}
		return nil
	}
}

func maxArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, got []string) error {
		if len(got) > n {
			return &UsageError{fmt.Errorf("%s takes at most %d argument(s)\nusage: %s", cmd.Name(), n, cmd.UseLine())}
		}
		return nil
	}
}

func (a *app) printJSON(v any) error {
	enc := json.NewEncoder(a.out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printGrouped prints aliases (already sorted by group, name) under group
// headings with names aligned across the whole listing.
func (a *app) printGrouped(list []alias.Alias) {
	width := 0
	for _, x := range list {
		if len(x.Name) > width {
			width = len(x.Name)
		}
	}
	prev := ""
	for i, x := range list {
		if x.Group != prev {
			if i > 0 {
				fmt.Fprintln(a.out)
			}
			fmt.Fprintln(a.out, x.Group)
			prev = x.Group
		}
		fmt.Fprintf(a.out, "  %-*s  %s\n", width, x.Name, display(x.Command))
	}
}

// display makes untrusted text safe to print: control characters (including
// ESC and the C1 range, which terminals treat as escape sequences) are shown as
// visible escapes instead of being interpreted.
func display(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == '\r':
			b.WriteString(`\r`)
		case r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f):
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func nonNil(l []alias.Alias) []alias.Alias {
	if l == nil {
		return []alias.Alias{}
	}
	return l
}
