package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/model"
	"github.com/josh-quinn/ali/internal/store"
)

func (a *app) subCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sub",
		Short: "Manage sub-aliases: short words that expand after a tool",
		Long: `A sub-alias is a short word that only means something after a given tool. With
"ali sub add kubectl g get" and "ali sub add kubectl p pods", typing "k g p" (where k
is an ali alias for kubectl) becomes "kubectl get pods" in your command line before it
runs. Helm can have its own "g".

Sub-aliases expand anywhere after the tool. Write \p to keep a word literal (for
example a pod called p); quoted words and anything after a bare -- are never expanded.

Shared sub-aliases (--shared) work after any tool ali knows (one with its own
sub-aliases, or one an ali alias runs). A tool's own sub-alias wins over a shared one
with the same name. See "ali help expansion".`,
	}
	cmd.AddCommand(a.subAddCmd(), a.subUpdateCmd(), a.subRmCmd(), a.subLsCmd())
	return cmd
}

// subTarget parses "TOOL NAME..." or "--shared NAME..." into a tool ("" = shared).
func subTarget(shared bool, args []string) (tool string, rest []string, err error) {
	if shared {
		return "", args, nil
	}
	if len(args) == 0 {
		return "", nil, &UsageError{fmt.Errorf("give a TOOL, or --shared")}
	}
	if err := model.ValidateTool(args[0]); err != nil {
		return "", nil, err
	}
	return args[0], args[1:], nil
}

// looseFlags parses args for commands whose values often start with "-" (sub-alias
// expansions such as "-o wide"). Only the exact flag words in known count as
// flags; everything else is positional. It returns help=true for -h/--help.
func looseFlags(args []string, known map[string]*bool) (pos []string, help bool) {
	for _, arg := range args {
		if p, ok := known[arg]; ok {
			*p = true
			continue
		}
		if arg == "-h" || arg == "--help" {
			return nil, true
		}
		pos = append(pos, arg)
	}
	return pos, false
}

func subLabel(tool, name string) string {
	if tool == "" {
		return "shared sub-alias " + name
	}
	return "sub-alias " + name + " for " + tool
}

func getSub(d *store.Data, tool, name string) (string, bool) {
	if tool == "" {
		exp, ok := d.Shared()[name]
		return exp, ok
	}
	exp, ok := d.Subs(tool)[name]
	return exp, ok
}

func setSub(d *store.Data, tool, name, exp string) {
	if tool == "" {
		d.SetShared(name, exp)
	} else {
		d.SetSub(tool, name, exp)
	}
}

func validateSub(name, exp string) error {
	if err := model.ValidateSubName(name); err != nil {
		return err
	}
	return alias.ValidateCommand(exp)
}

func (a *app) subAddCmd() *cobra.Command {
	var shared, force bool
	cmd := &cobra.Command{
		Use:   "add (TOOL | --shared) NAME EXPANSION [--force]",
		Short: "Add a sub-alias for a tool, or a shared one",
		Long: `Add sub-alias NAME that expands to EXPANSION after TOOL (the program's name, e.g.
kubectl), or after any known tool with --shared. EXPANSION may be several words and may
contain {{name}} placeholders for variables (see "ali help placeholders"). Fails if the
sub-alias exists, unless --force is given.`,
		Example: `  ali sub add kubectl g get
  ali sub add kubectl p pods
  ali sub add helm g 'get values'
  ali sub add --shared o '-o wide'`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, raw []string) error {
			args, help := looseFlags(raw, map[string]*bool{"--shared": &shared, "--force": &force})
			if help {
				return cmd.Help()
			}
			if want := map[bool]int{true: 2, false: 3}[shared]; len(args) != want {
				return &UsageError{fmt.Errorf("usage: %s", cmd.UseLine())}
			}
			tool, rest, err := subTarget(shared, args)
			if err != nil {
				return err
			}
			name, exp := rest[0], rest[1]
			if err := validateSub(name, exp); err != nil {
				return err
			}
			return a.st.Update(func(d *store.Data) error {
				if err := a.checkPlaceholders(d, exp); err != nil {
					return err
				}
				if prev, ok := getSub(d, tool, name); ok && !force {
					return fmt.Errorf("%s already exists (%s); use --force or ali sub update", subLabel(tool, name), display(prev))
				}
				setSub(d, tool, name, exp)
				fmt.Fprintf(a.out, "added %s: %s\n", subLabel(tool, name), display(exp))
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&shared, "shared", false, "add a shared sub-alias usable after any known tool")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing sub-alias")
	return cmd
}

func (a *app) subUpdateCmd() *cobra.Command {
	var shared bool
	cmd := &cobra.Command{
		Use:                "update (TOOL | --shared) NAME EXPANSION",
		Short:              "Change what a sub-alias expands to",
		Long:               `Replace the expansion of an existing sub-alias.`,
		Example:            `  ali sub update kubectl p 'pods -o wide'`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, raw []string) error {
			args, help := looseFlags(raw, map[string]*bool{"--shared": &shared})
			if help {
				return cmd.Help()
			}
			if want := map[bool]int{true: 2, false: 3}[shared]; len(args) != want {
				return &UsageError{fmt.Errorf("usage: %s", cmd.UseLine())}
			}
			tool, rest, err := subTarget(shared, args)
			if err != nil {
				return err
			}
			name, exp := rest[0], rest[1]
			if err := validateSub(name, exp); err != nil {
				return err
			}
			return a.st.Update(func(d *store.Data) error {
				if _, ok := getSub(d, tool, name); !ok {
					return fmt.Errorf("no %s", subLabel(tool, name))
				}
				if err := a.checkPlaceholders(d, exp); err != nil {
					return err
				}
				setSub(d, tool, name, exp)
				fmt.Fprintf(a.out, "updated %s: %s\n", subLabel(tool, name), display(exp))
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&shared, "shared", false, "update a shared sub-alias")
	return cmd
}

func (a *app) subRmCmd() *cobra.Command {
	var shared bool
	cmd := &cobra.Command{
		Use:   "rm (TOOL | --shared) NAME...",
		Short: "Remove sub-aliases",
		Long:  `Remove one or more sub-aliases of TOOL, or shared ones with --shared. Names that do not exist are reported and make the exit status 1; the others are still removed.`,
		Example: `  ali sub rm kubectl g p
  ali sub rm --shared o`,
		Args: minArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tool, names, err := subTarget(shared, args)
			if err != nil {
				return err
			}
			if len(names) == 0 {
				return &UsageError{fmt.Errorf("give at least one NAME")}
			}
			var missing []string
			if err := a.st.Update(func(d *store.Data) error {
				for _, n := range names {
					exp, ok := getSub(d, tool, n)
					if !ok {
						missing = append(missing, n)
						continue
					}
					if tool == "" {
						d.DeleteShared(n)
					} else {
						d.DeleteSub(tool, n)
					}
					fmt.Fprintf(a.out, "removed %s: %s\n", subLabel(tool, n), display(exp))
				}
				return nil
			}); err != nil {
				return err
			}
			if len(missing) > 0 {
				return fmt.Errorf("no such sub-alias: %s", strings.Join(missing, ", "))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&shared, "shared", false, "remove shared sub-aliases")
	return cmd
}

type subRow struct {
	Tool      string `json:"tool,omitempty"`
	Name      string `json:"name"`
	Expansion string `json:"expansion"`
	Shared    bool   `json:"shared,omitempty"`
}

func (a *app) subLsCmd() *cobra.Command {
	var shared, asJSON bool
	cmd := &cobra.Command{
		Use:   "ls [TOOL | --shared] [--json]",
		Short: "List sub-aliases, grouped by tool",
		Long:  `List sub-aliases under their tool, with shared ones last. Give a TOOL to see only that tool's, or --shared for only the shared ones. "ali show TOOL" also lists the aliases and variables that go with a tool.`,
		Example: `  ali sub ls
  ali sub ls kubectl
  ali sub ls --json`,
		Args: maxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := a.st.Load()
			if err != nil {
				return err
			}
			var rows []subRow
			addTool := func(tool string) {
				subs := d.Subs(tool)
				for _, n := range sortedNames(subs) {
					rows = append(rows, subRow{Tool: tool, Name: n, Expansion: subs[n]})
				}
			}
			switch {
			case shared:
			case len(args) == 1:
				if len(d.Subs(args[0])) == 0 {
					return fmt.Errorf("no sub-aliases for %q (see: ali sub ls)", args[0])
				}
				addTool(args[0])
			default:
				for _, tool := range d.Tools() {
					addTool(tool)
				}
			}
			if shared || len(args) == 0 {
				sh := d.Shared()
				for _, n := range sortedNames(sh) {
					rows = append(rows, subRow{Name: n, Expansion: sh[n], Shared: true})
				}
			}
			if asJSON {
				if rows == nil {
					rows = []subRow{}
				}
				return a.printJSON(rows)
			}
			if len(rows) == 0 {
				fmt.Fprintln(a.out, "no sub-aliases yet. try: ali sub add kubectl g get")
				return nil
			}
			a.printSubs(rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&shared, "shared", false, "list only shared sub-aliases")
	cmd.Flags().BoolVar(&asJSON, "json", false, "machine-readable output")
	return cmd
}

func (a *app) printSubs(rows []subRow) {
	width := 0
	for _, r := range rows {
		if len(r.Name) > width {
			width = len(r.Name)
		}
	}
	prev := "\x00"
	for i, r := range rows {
		head := r.Tool
		if r.Shared {
			head = "shared"
		}
		if head != prev {
			if i > 0 {
				fmt.Fprintln(a.out)
			}
			fmt.Fprintln(a.out, head)
			prev = head
		}
		fmt.Fprintf(a.out, "  %-*s  %s\n", width, r.Name, display(r.Expansion))
	}
}

func sortedNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
