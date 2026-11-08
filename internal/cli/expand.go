package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/expand"
	"github.com/josh-quinn/ali/internal/model"
	"github.com/josh-quinn/ali/internal/settings"
	"github.com/josh-quinn/ali/internal/store"
)

// defsFrom builds the expansion definitions from the store.
func defsFrom(d *store.Data) *expand.Defs {
	defs := &expand.Defs{
		Aliases: map[string]string{},
		Subs:    map[string]map[string]string{},
		Shared:  d.Shared(),
		Vars:    map[string]model.Variable{},
		Envs:    map[string]model.Environment{},
	}
	for _, x := range d.All() {
		defs.Aliases[x.Name] = x.Command
	}
	for _, tool := range d.Tools() {
		defs.Subs[tool] = d.Subs(tool)
	}
	for _, n := range d.VarNames() {
		defs.Vars[n], _ = d.Var(n)
	}
	for _, n := range d.EnvNames() {
		defs.Envs[n], _ = d.Env(n)
	}
	return defs
}

// expandOptions reads the terminal and global environment choices.
func (a *app) expandOptions(mode expand.Mode) expand.Options {
	global, _ := settings.GlobalEnv(a.st.Dir)
	return expand.Options{Mode: mode, TerminalEnv: os.Getenv("ALI_ENV"), GlobalEnv: global}
}

func (a *app) expandCmd() *cobra.Command {
	var zle string
	var run bool
	cmd := &cobra.Command{
		Use:   "expand LINE",
		Short: "Show what a command line expands to, without running it",
		Long: `Print the command that LINE would run after ali expands its sub-aliases,
placeholders and @ values, using the current terminal's environment. Nothing is run.
Exits with status 1 and a message if a placeholder has no value or an @ value is
ambiguous. Quote LINE so your shell passes it as one argument.

The ali shell integration also uses this command internally (--zle).`,
		Example: `  ali expand 'k g p'
  ali expand 'k g p n @sys'`,
		Args: func(c *cobra.Command, args []string) error {
			if run && len(args) > 0 {
				return nil
			}
			if len(args) != 1 {
				return &UsageError{fmt.Errorf("give one LINE (quote it)\nusage: %s", c.UseLine())}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if zle != "" {
				return a.expandZle(zle, args[0])
			}
			if run {
				return a.expandRun(args)
			}
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			res, err := expand.Expand(args[0], defsFrom(d), a.expandOptions(expand.Full))
			for _, w := range res.Warnings {
				fmt.Fprintln(a.err, "warning:", w)
			}
			if err != nil {
				return err
			}
			fmt.Fprintln(a.out, res.Line)
			return nil
		},
	}
	cmd.Flags().StringVar(&zle, "zle", "", "internal: protocol for the zsh widgets (space or accept)")
	cmd.Flags().MarkHidden("zle")
	cmd.Flags().BoolVar(&run, "run", false, "internal: expand already-split words for eval (alias functions)")
	cmd.Flags().MarkHidden("run")
	return cmd
}

// expandZle answers the line-editor widgets (research R4) with NUL-separated
// fields: status (ok, unchanged, confirm, error), new text, message. It never
// returns an error itself, so a widget can always fall back to plain zsh.
func (a *app) expandZle(kind, text string) error {
	var mode expand.Mode
	switch kind {
	case "space":
		mode = expand.Partial
	case "accept":
		mode = expand.Full
	default:
		return &UsageError{fmt.Errorf("--zle must be space or accept, not %q", kind)}
	}
	reply := func(status, line, msg string) error {
		fmt.Fprint(a.out, strings.Join([]string{status, line, msg}, "\x00"))
		return nil
	}
	d, err := a.st.Load()
	if err != nil {
		return reply("error", text, "ali: "+err.Error())
	}
	res, err := expand.Expand(text, defsFrom(d), a.expandOptions(mode))
	msg := strings.Join(res.Warnings, "; ")
	switch {
	case err != nil:
		return reply("error", text, "ali: "+err.Error())
	case res.Protected:
		return reply("confirm", res.Line, fmt.Sprintf("ali: environment %q is protected; press Enter again to run", res.Env))
	case !res.Changed:
		return reply("unchanged", text, msg)
	}
	return reply("ok", res.Line, msg)
}

var (
	literalWordRe = regexp.MustCompile(`^\\[A-Za-z0-9_.@+:=,%/-]+$`)
	atAssignRe    = regexp.MustCompile(`^@[a-z][a-z0-9_]*=`)
)

// expandRun serves alias functions (research R5): the shell has already split
// the words, so each is re-quoted before expansion, and the result is printed
// for eval. Every word in the output is ali-quoted.
func (a *app) expandRun(words []string) error {
	parts := []string{words[0]}
	for _, w := range words[1:] {
		switch {
		case literalWordRe.MatchString(w): // '\p' typed in quotes: keep it literal
			parts = append(parts, w)
		case atAssignRe.MatchString(w):
			eq := strings.IndexByte(w, '=')
			parts = append(parts, w[:eq+1]+expand.QuoteValue(w[eq+1:]))
		default:
			parts = append(parts, expand.QuoteValue(w))
		}
	}
	d, err := a.st.Load()
	if err != nil {
		return err
	}
	opt := a.expandOptions(expand.Full)
	opt.ForceAlias = true
	res, err := expand.Expand(strings.Join(parts, " "), defsFrom(d), opt)
	for _, w := range res.Warnings {
		fmt.Fprintln(a.err, "warning:", w)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(a.out, res.Line)
	return nil
}
