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
		Use:           "ali",
		Short:         "Create, find and organise shell aliases without touching .zshrc",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(out)
	root.SetErr(errw)
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return &UsageError{err} })
	root.CompletionOptions.DisableDefaultCmd = true
	root.AddCommand(a.addCmd(), a.lsCmd(), a.groupsCmd(), a.searchCmd(), a.regroupCmd(), a.updateCmd(), a.rmCmd())
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

func display(cmd string) string { return strings.ReplaceAll(cmd, "\n", `\n`) }

func nonNil(l []alias.Alias) []alias.Alias {
	if l == nil {
		return []alias.Alias{}
	}
	return l
}
