package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/model"
	"github.com/josh-quinn/ali/internal/settings"
	"github.com/josh-quinn/ali/internal/store"
)

func (a *app) envCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environments: sets of variable values you switch between",
		Long: `An environment sets variables for every command at once. With
"ali env add dev ns=one ctx=dev" and "ali env add prod ns=prod ctx=prod --protect",
"ali env use prod" makes {{ns}} and {{ctx}} resolve from prod in this terminal only;
"--global" chooses for every terminal that has not chosen its own.

The active environment is shown in your prompt. Commands that take values from a
protected environment are shown in full and need a second Enter before they run.
Use @@name to use another environment for a single command. See
"ali help environments".`,
	}
	cmd.AddCommand(a.envAddCmd(), a.envSetCmd(), a.envUnsetCmd(), a.envProtectCmd(true), a.envProtectCmd(false),
		a.envRmCmd(), a.envLsCmd(), a.envShowCmd(), a.envUseCmd(), a.envOffCmd(), a.envCurrentCmd())
	return cmd
}

func (a *app) parseEnvSet(d *store.Data, args []string) (map[string]string, error) {
	set, err := parseKV(args, func(k string) error {
		if _, ok := d.Var(k); !ok {
			return fmt.Errorf("no variable %q (create it with: ali var add %s ...)", k, k)
		}
		return nil
	})
	return set, err
}

func (a *app) withEnv(name string, fn func(d *store.Data, e *model.Environment) error) error {
	return a.st.Update(func(d *store.Data) error {
		e, ok := d.Env(name)
		if !ok {
			return fmt.Errorf("no environment %q (see: ali env ls)", name)
		}
		if e.Set == nil {
			e.Set = map[string]string{}
		}
		if err := fn(d, &e); err != nil {
			return err
		}
		d.SetEnv(name, e)
		return nil
	})
}

func (a *app) envAddCmd() *cobra.Command {
	var protect bool
	cmd := &cobra.Command{
		Use:   "add NAME [VAR=VALUE...] [--protect]",
		Short: "Create an environment",
		Long: `Create environment NAME setting each VAR to VALUE. VALUE can be one of the
variable's value names (e.g. one) or literal text. --protect makes commands that use
the environment's values wait for a second Enter.`,
		Example: `  ali env add dev ns=one ctx=dev
  ali env add prod ns=production ctx=prod --protect`,
		Args: minArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := model.ValidateEnvName(args[0]); err != nil {
				return err
			}
			return a.st.Update(func(d *store.Data) error {
				if _, ok := d.Env(args[0]); ok {
					return fmt.Errorf("environment %q already exists; use ali env set", args[0])
				}
				set, err := a.parseEnvSet(d, args[1:])
				if err != nil {
					return err
				}
				d.SetEnv(args[0], model.Environment{Protected: protect, Set: set})
				fmt.Fprintf(a.out, "added environment %s\n", args[0])
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&protect, "protect", false, "require a second Enter for commands using this environment")
	return cmd
}

func (a *app) envSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set NAME VAR=VALUE...",
		Short:   "Set variables in an environment",
		Long:    `Add or change variable assignments in environment NAME.`,
		Example: `  ali env set dev ns=sys`,
		Args:    minArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.withEnv(args[0], func(d *store.Data, e *model.Environment) error {
				set, err := a.parseEnvSet(d, args[1:])
				if err != nil {
					return err
				}
				for k, v := range set {
					e.Set[k] = v
				}
				fmt.Fprintf(a.out, "updated environment %s\n", args[0])
				return nil
			})
		},
	}
}

func (a *app) envUnsetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "unset NAME VAR...",
		Short:   "Remove variables from an environment",
		Long:    `Stop environment NAME from setting the given variables.`,
		Example: `  ali env unset dev ctx`,
		Args:    minArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.withEnv(args[0], func(_ *store.Data, e *model.Environment) error {
				for _, v := range args[1:] {
					if _, ok := e.Set[v]; !ok {
						return fmt.Errorf("environment %s does not set %q", args[0], v)
					}
					delete(e.Set, v)
				}
				fmt.Fprintf(a.out, "updated environment %s\n", args[0])
				return nil
			})
		},
	}
}

func (a *app) envProtectCmd(on bool) *cobra.Command {
	use, short, long := "protect NAME", "Require a second Enter for commands using an environment",
		`Mark environment NAME as protected: a command that takes a value from it is shown in full and runs only after a second Enter (or a confirmation in run-time mode).`
	if !on {
		use, short, long = "unprotect NAME", "Stop requiring a second Enter for an environment", `Remove the protection from environment NAME.`
	}
	return &cobra.Command{
		Use:     use,
		Short:   short,
		Long:    long,
		Example: "  ali env " + strings.Fields(use)[0] + " prod",
		Args:    args(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.withEnv(args[0], func(_ *store.Data, e *model.Environment) error {
				e.Protected = on
				fmt.Fprintf(a.out, "updated environment %s\n", args[0])
				return nil
			})
		},
	}
}

func (a *app) envRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm NAME",
		Short: "Remove an environment",
		Long: `Remove environment NAME. If it was the global choice, that choice is cleared.
Terminals that chose it themselves fall back to no environment, with a warning, the
next time a command needs it.`,
		Example: `  ali env rm staging`,
		Args:    args(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.st.Update(func(d *store.Data) error {
				if !d.DeleteEnv(args[0]) {
					return fmt.Errorf("no environment %q", args[0])
				}
				fmt.Fprintf(a.out, "removed environment %s\n", args[0])
				return nil
			}); err != nil {
				return err
			}
			if g, _ := settings.GlobalEnv(a.st.Dir); g == args[0] {
				if err := settings.SetGlobalEnv(a.st.Dir, ""); err != nil {
					return err
				}
				fmt.Fprintf(a.err, "note: %s was the global environment; there is no global environment now\n", args[0])
			}
			return nil
		},
	}
}

// active returns the environment in effect and where it comes from.
func (a *app) active() (name, from string) {
	if t := os.Getenv("ALI_ENV"); t != "" {
		return t, "this terminal"
	}
	if g, _ := settings.GlobalEnv(a.st.Dir); g != "" {
		return g, "global"
	}
	return "", ""
}

type envRow struct {
	Name      string            `json:"name"`
	Protected bool              `json:"protected,omitempty"`
	Active    bool              `json:"active,omitempty"`
	Set       map[string]string `json:"set"`
}

func (a *app) envLsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "ls [--json]",
		Short: "List environments; the active one is marked",
		Long:  `List every environment. "*" marks the one active in this terminal.`,
		Example: `  ali env ls
  ali env ls --json`,
		Args: args(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			act, _ := a.active()
			rows := []envRow{}
			for _, n := range d.EnvNames() {
				e, _ := d.Env(n)
				rows = append(rows, envRow{Name: n, Protected: e.Protected, Active: n == act, Set: e.Set})
			}
			if asJSON {
				return a.printJSON(rows)
			}
			if len(rows) == 0 {
				fmt.Fprintln(a.out, "no environments yet. try: ali env add dev ns=one")
			}
			for _, r := range rows {
				mark, prot := " ", ""
				if r.Active {
					mark = "*"
				}
				if r.Protected {
					prot = "  (protected)"
				}
				fmt.Fprintf(a.out, "%s %s%s\n", mark, r.Name, prot)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}

func (a *app) envShowCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show [NAME] [--json]",
		Short: "Show what an environment sets",
		Long:  `Show each variable environment NAME sets (the active one if NAME is omitted), with value names resolved to their values.`,
		Example: `  ali env show prod
  ali env show`,
		Args: maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			name, _ := a.active()
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				return fmt.Errorf("no environment is active; give a NAME (see: ali env ls)")
			}
			e, ok := d.Env(name)
			if !ok {
				return fmt.Errorf("no environment %q (see: ali env ls)", name)
			}
			if asJSON {
				return a.printJSON(envRow{Name: name, Protected: e.Protected, Set: e.Set})
			}
			prot := ""
			if e.Protected {
				prot = " (protected)"
			}
			fmt.Fprintln(a.out, name+prot)
			width := 0
			for k := range e.Set {
				if len(k) > width {
					width = len(k)
				}
			}
			for _, k := range sortedNames(e.Set) {
				shown := display(e.Set[k])
				if v, _ := d.Var(k); v.Values[e.Set[k]] != "" {
					shown += " (" + display(v.Values[e.Set[k]]) + ")"
				}
				fmt.Fprintf(a.out, "  %-*s  %s\n", width, k, shown)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}

func (a *app) envUseCmd() *cobra.Command {
	var global bool
	cmd := &cobra.Command{
		Use:   "use NAME [--global]",
		Short: "Choose the environment for this terminal, or for all terminals",
		Long: `Use environment NAME in this terminal only. With --global, use it in every
terminal that has not chosen its own (including this one, whose own choice is
cleared). The choice for this terminal is made by the ali shell function, so run
this from your shell.`,
		Example: `  ali env use prod
  ali env use dev --global`,
		Args: args(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			if _, ok := d.Env(args[0]); !ok {
				return fmt.Errorf("no environment %q (see: ali env ls)", args[0])
			}
			if global {
				if err := settings.SetGlobalEnv(a.st.Dir, args[0]); err != nil {
					return err
				}
				fmt.Fprintf(a.out, "using %s in every terminal that has not chosen its own\n", args[0])
				return nil
			}
			if os.Getenv("ALI_WRAPPER") == "" {
				fmt.Fprintf(a.err, "note: only your shell can change this terminal's environment; run 'ali env use %s' at your prompt (with ali's block installed), or use --global\n", args[0])
				return nil
			}
			fmt.Fprintf(a.out, "using %s in this terminal\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&global, "global", false, "choose for all terminals that have not chosen their own")
	return cmd
}

func (a *app) envOffCmd() *cobra.Command {
	var global bool
	cmd := &cobra.Command{
		Use:   "off [--global]",
		Short: "Stop using an environment",
		Long:  `Stop using an environment in this terminal (it then follows the global choice, if any). With --global, clear the global choice.`,
		Example: `  ali env off
  ali env off --global`,
		Args: args(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if global {
				if err := settings.SetGlobalEnv(a.st.Dir, ""); err != nil {
					return err
				}
				fmt.Fprintln(a.out, "no global environment")
				return nil
			}
			fmt.Fprintln(a.out, "this terminal no longer chooses its own environment")
			return nil
		},
	}
	cmd.Flags().BoolVar(&global, "global", false, "clear the global choice")
	return cmd
}

func (a *app) envCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "current",
		Short:   "Show which environment is active here and why",
		Long:    `Print the environment in effect for this terminal and where it comes from: "this terminal", "global", or none.`,
		Example: `  ali env current`,
		Args:    args(0),
		RunE: func(cmd *cobra.Command, _ []string) error {
			name, from := a.active()
			if name == "" {
				fmt.Fprintln(a.out, "none")
				return nil
			}
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			e, ok := d.Env(name)
			switch {
			case !ok:
				from += ", no longer exists"
			case e.Protected:
				from += ", protected"
			}
			fmt.Fprintf(a.out, "%s (%s)\n", name, from)
			return nil
		},
	}
}
