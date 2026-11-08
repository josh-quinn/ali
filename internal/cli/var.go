package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/model"
	"github.com/josh-quinn/ali/internal/store"
)

func (a *app) varCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "var",
		Short: "Manage variables: named values for {{name}} placeholders",
		Long: `A variable gives names to the values a placeholder can take. With
"ali var add ns one=namespace-1 sys=kube-system --default one" and a sub-alias
"ali sub add kubectl n '-n {{ns}}'", typing "k g p n @sys" runs
"kubectl get pods -n kube-system", and "k g p n" uses the default.

Values are data: they are inserted as one quoted word, never run as shell code.
Environments (ali env) can set a variable for a whole terminal. See
"ali help placeholders".`,
	}
	cmd.AddCommand(a.varAddCmd(), a.varSetCmd(), a.varUnsetCmd(), a.varDefaultCmd(), a.varRmCmd(), a.varLsCmd(), a.varShowCmd())
	return cmd
}

// parseKV parses KEY=VALUE arguments for a variable's values.
func parseKV(args []string, validKey func(string) error) (map[string]string, error) {
	out := map[string]string{}
	for _, arg := range args {
		k, v, ok := strings.Cut(arg, "=")
		if !ok {
			return nil, &UsageError{fmt.Errorf("%q is not KEY=VALUE", arg)}
		}
		if err := validKey(k); err != nil {
			return nil, err
		}
		if err := model.ValidateValue(v); err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		out[k] = v
	}
	return out, nil
}

func (a *app) withVar(name string, fn func(d *store.Data, v *model.Variable) error) error {
	return a.st.Update(func(d *store.Data) error {
		v, ok := d.Var(name)
		if !ok {
			return fmt.Errorf("no variable %q (see: ali var ls)", name)
		}
		if err := fn(d, &v); err != nil {
			return err
		}
		if err := v.Validate(); err != nil {
			return err
		}
		d.SetVar(name, v)
		return nil
	})
}

func (a *app) varAddCmd() *cobra.Command {
	var def string
	cmd := &cobra.Command{
		Use:   "add NAME [KEY=VALUE...] [--default KEY]",
		Short: "Create a variable with named values",
		Long: `Create variable NAME. Each KEY=VALUE gives a value a short name, which you pick
on the command line with @KEY. --default chooses the value used when nothing else
sets the variable.`,
		Example: `  ali var add ns one=namespace-1 sys=kube-system --default one
  ali var add ctx dev=dev-cluster prod=prod-cluster`,
		Args: minArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			name := args[0]
			if err := model.ValidateVarName(name); err != nil {
				return err
			}
			vals, err := parseKV(args[1:], model.ValidateKey)
			if err != nil {
				return err
			}
			v := model.Variable{Values: vals, Default: def}
			if err := v.Validate(); err != nil {
				return err
			}
			return a.st.Update(func(d *store.Data) error {
				if _, ok := d.Var(name); ok {
					return fmt.Errorf("variable %q already exists; use ali var set", name)
				}
				d.SetVar(name, v)
				fmt.Fprintf(a.out, "added variable %s (%d value(s))\n", name, len(vals))
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&def, "default", "", "value KEY to use when nothing else sets the variable")
	return cmd
}

func (a *app) varSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set NAME KEY=VALUE...",
		Short:   "Add or change values of a variable",
		Long:    `Add new named values to variable NAME, or change existing ones.`,
		Example: `  ali var set ns dev=dev-namespace`,
		Args:    minArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			vals, err := parseKV(args[1:], model.ValidateKey)
			if err != nil {
				return err
			}
			return a.withVar(args[0], func(_ *store.Data, v *model.Variable) error {
				for k, val := range vals {
					v.Values[k] = val
				}
				fmt.Fprintf(a.out, "updated variable %s\n", args[0])
				return nil
			})
		},
	}
}

func (a *app) varUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "unset NAME KEY...",
		Short:   "Remove values from a variable",
		Long:    `Remove named values from variable NAME. Removing the default value also clears the default.`,
		Example: `  ali var unset ns dev`,
		Args:    minArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return a.withVar(args[0], func(_ *store.Data, v *model.Variable) error {
				for _, k := range args[1:] {
					if _, ok := v.Values[k]; !ok {
						return fmt.Errorf("variable %s has no value %q", args[0], k)
					}
					delete(v.Values, k)
					if v.Default == k {
						v.Default = ""
						fmt.Fprintf(a.err, "note: %s was the default; %s now has no default\n", k, args[0])
					}
				}
				fmt.Fprintf(a.out, "updated variable %s\n", args[0])
				return nil
			})
		},
	}
}

func (a *app) varDefaultCmd() *cobra.Command {
	var none bool
	cmd := &cobra.Command{
		Use:   "default NAME (KEY | --none)",
		Short: "Set or clear a variable's default value",
		Long:  `Choose which named value NAME uses when no @value or environment sets it, or clear the default with --none.`,
		Example: `  ali var default ns one
  ali var default ns --none`,
		Args: func(c *cobra.Command, args []string) error {
			if want := map[bool]int{true: 1, false: 2}[none]; len(args) != want {
				return &UsageError{fmt.Errorf("usage: %s", c.UseLine())}
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return a.withVar(args[0], func(_ *store.Data, v *model.Variable) error {
				if none {
					v.Default = ""
				} else {
					if _, ok := v.Values[args[1]]; !ok {
						return fmt.Errorf("variable %s has no value %q", args[0], args[1])
					}
					v.Default = args[1]
				}
				fmt.Fprintf(a.out, "updated variable %s\n", args[0])
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&none, "none", false, "clear the default")
	return cmd
}

func (a *app) varRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "rm NAME",
		Short:   "Remove a variable that nothing uses",
		Long:    `Remove variable NAME. This is refused while any alias, sub-alias or environment still uses it; the message lists them.`,
		Example: `  ali var rm ns`,
		Args:    args(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return a.st.Update(func(d *store.Data) error {
				if _, ok := d.Var(args[0]); !ok {
					return fmt.Errorf("no variable %q", args[0])
				}
				if uses := d.UsesOf(args[0]); len(uses) > 0 {
					return fmt.Errorf("variable %s is still used by: %s", args[0], strings.Join(uses, ", "))
				}
				d.DeleteVar(args[0])
				fmt.Fprintf(a.out, "removed variable %s\n", args[0])
				return nil
			})
		},
	}
}

type varRow struct {
	Name    string            `json:"name"`
	Default string            `json:"default,omitempty"`
	Values  map[string]string `json:"values"`
	UsedBy  []string          `json:"used_by"`
}

func (a *app) varRows(d *store.Data, names []string) []varRow {
	rows := []varRow{}
	for _, n := range names {
		v, _ := d.Var(n)
		uses := d.UsesOf(n)
		if uses == nil {
			uses = []string{}
		}
		rows = append(rows, varRow{Name: n, Default: v.Default, Values: v.Values, UsedBy: uses})
	}
	return rows
}

func (a *app) printVar(r varRow, detail bool) {
	fmt.Fprintln(a.out, r.Name)
	width := 0
	for k := range r.Values {
		if len(k) > width {
			width = len(k)
		}
	}
	for _, k := range sortedNames(r.Values) {
		mark := ""
		if k == r.Default {
			mark = "  (default)"
		}
		fmt.Fprintf(a.out, "  %-*s  %s%s\n", width, k, display(r.Values[k]), mark)
	}
	if detail {
		if len(r.UsedBy) == 0 {
			fmt.Fprintln(a.out, "  used by: nothing yet")
		} else {
			fmt.Fprintf(a.out, "  used by: %s\n", strings.Join(r.UsedBy, ", "))
		}
	}
}

func (a *app) varLsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "ls [--json]",
		Short: "List variables and their values",
		Long:  `List every variable with its named values; the default is marked.`,
		Example: `  ali var ls
  ali var ls --json`,
		Args: args(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			rows := a.varRows(d, d.VarNames())
			if asJSON {
				return a.printJSON(rows)
			}
			if len(rows) == 0 {
				fmt.Fprintln(a.out, "no variables yet. try: ali var add ns one=namespace-1 --default one")
			}
			for i, r := range rows {
				if i > 0 {
					fmt.Fprintln(a.out)
				}
				a.printVar(r, false)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}

func (a *app) varShowCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "show NAME [--json]",
		Short:   "Show a variable's values and everything that uses it",
		Long:    `Show variable NAME's values (default marked) and the aliases, sub-aliases and environments that use it.`,
		Example: `  ali var show ns`,
		Args:    args(1),
		RunE: func(_ *cobra.Command, args []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			if _, ok := d.Var(args[0]); !ok {
				return fmt.Errorf("no variable %q (see: ali var ls)", args[0])
			}
			r := a.varRows(d, args[:1])[0]
			if asJSON {
				return a.printJSON(r)
			}
			a.printVar(r, true)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}
