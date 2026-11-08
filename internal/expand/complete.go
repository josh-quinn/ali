package expand

import (
	"sort"
	"strings"

	"github.com/josh-quinn/ali/internal/model"
)

// Complete returns Tab-completion candidates for the word being typed at the
// end of lbuf (FR-016): "@" value names (and "@var=" forms when several
// variables are open) for placeholders in the current command, "@@" environment
// names, or the tool's sub-aliases. nil means "not ours": the shell's own
// completion should handle it.
func Complete(lbuf string, d *Defs) []string {
	cur := lbuf[strings.LastIndexAny(lbuf, " \t")+1:]
	prefix := lbuf[:len(lbuf)-len(cur)]
	exp, _ := Expand(prefix, d, Options{Mode: Partial})
	cmds := Tokenize(exp.Line)
	if len(cmds) == 0 || strings.HasSuffix(strings.TrimRight(prefix, " \t"), ";") || strings.HasSuffix(strings.TrimRight(prefix, " \t"), "|") {
		return nil
	}
	last := cmds[len(cmds)-1]
	info := d.toolOf(last)
	if info.cmdWord < 0 {
		return nil
	}
	set := map[string]bool{}
	switch {
	case strings.HasPrefix(cur, "@@"):
		if placeholderVars(last, d) == nil {
			return nil
		}
		for n := range d.Envs {
			set["@@"+n] = true
		}
	case strings.HasPrefix(cur, "@"):
		vars := placeholderVars(last, d)
		for _, v := range vars {
			for k := range d.Vars[v].Values {
				set["@"+k] = true
			}
			if len(vars) > 1 {
				set["@"+v+"="] = true
			}
		}
	default:
		if info.tool == "" {
			return nil
		}
		for n := range d.Subs[info.tool] {
			set[n] = true
		}
		if d.knownTool(info.tool) {
			for n := range d.Shared {
				set[n] = true
			}
		}
	}
	var out []string
	for c := range set {
		if strings.HasPrefix(c, cur) {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

// placeholderVars lists defined variables with placeholders in a command.
func placeholderVars(c Command, d *Defs) []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range c.Words {
		if !w.Plain {
			continue
		}
		for _, n := range model.PlaceholderNames(w.Text) {
			if _, ok := d.Vars[n]; ok && !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	sort.Strings(out)
	return out
}
