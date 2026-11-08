package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/group"
	"github.com/josh-quinn/ali/internal/model"
)

type showOut struct {
	Tool       string            `json:"tool"`
	Aliases    []alias.Alias     `json:"aliases"`
	Subs       map[string]string `json:"sub_aliases"`
	Shared     map[string]string `json:"shared_sub_aliases"`
	Overridden []string          `json:"shared_overridden"`
	Variables  []string          `json:"variables"`
}

func (a *app) showCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show TOOL [--json]",
		Short: "Show everything you can use with a tool",
		Long: `Show, for TOOL (e.g. kubectl): the ali aliases that run it, its sub-aliases, the
shared sub-aliases that apply to it (marked when the tool's own one wins), and the
variables those use. Use "ali var show NAME" for a variable's values.`,
		Example: `  ali show kubectl
  ali show helm --json`,
		Args: args(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			tool := args[0]
			o := showOut{Tool: tool, Aliases: []alias.Alias{}, Subs: d.Subs(tool), Shared: map[string]string{}, Overridden: []string{}, Variables: []string{}}
			for _, x := range d.All() {
				if group.Infer(x.Command) == tool {
					o.Aliases = append(o.Aliases, x)
				}
			}
			if len(o.Aliases) == 0 && len(o.Subs) == 0 {
				return fmt.Errorf("ali knows nothing about %q (no aliases run it and it has no sub-aliases)", tool)
			}
			o.Shared = d.Shared()
			for n := range o.Shared {
				if _, ok := o.Subs[n]; ok {
					o.Overridden = append(o.Overridden, n)
				}
			}
			sort.Strings(o.Overridden)
			vars := map[string]bool{}
			collect := func(text string) {
				for _, p := range model.PlaceholderNames(text) {
					if _, ok := d.Var(p); ok {
						vars[p] = true
					}
				}
			}
			for _, x := range o.Aliases {
				collect(x.Command)
			}
			for _, exp := range o.Subs {
				collect(exp)
			}
			for _, exp := range o.Shared {
				collect(exp)
			}
			for v := range vars {
				o.Variables = append(o.Variables, v)
			}
			sort.Strings(o.Variables)
			if asJSON {
				return a.printJSON(o)
			}
			a.printShow(o)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}

func (a *app) printShow(o showOut) {
	section := func(title string, rows [][2]string) {
		if len(rows) == 0 {
			return
		}
		fmt.Fprintln(a.out, title)
		w := 0
		for _, r := range rows {
			if len(r[0]) > w {
				w = len(r[0])
			}
		}
		for _, r := range rows {
			fmt.Fprintf(a.out, "  %-*s  %s\n", w, r[0], r[1])
		}
		fmt.Fprintln(a.out)
	}
	var al [][2]string
	for _, x := range o.Aliases {
		al = append(al, [2]string{x.Name, display(x.Command)})
	}
	section("aliases", al)
	var subs [][2]string
	for _, n := range sortedNames(o.Subs) {
		subs = append(subs, [2]string{n, display(o.Subs[n])})
	}
	section("sub-aliases", subs)
	over := map[string]bool{}
	for _, n := range o.Overridden {
		over[n] = true
	}
	var sh [][2]string
	for _, n := range sortedNames(o.Shared) {
		v := display(o.Shared[n])
		if over[n] {
			v += fmt.Sprintf("  (overridden by %s's %s)", o.Tool, n)
		}
		sh = append(sh, [2]string{n, v})
	}
	section("shared sub-aliases", sh)
	if len(o.Variables) > 0 {
		fmt.Fprintln(a.out, "variables")
		for _, v := range o.Variables {
			fmt.Fprintf(a.out, "  %s   (ali var show %s)\n", v, v)
		}
	}
}
