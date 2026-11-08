package expand

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/josh-quinn/ali/internal/model"
)

var (
	atEnvRe = regexp.MustCompile(`^@@([a-z][a-z0-9_-]*)$`)
	atSetRe = regexp.MustCompile(`^@([a-z][a-z0-9_]*)=`)
	atKeyRe = regexp.MustCompile(`^@([a-z0-9][a-z0-9_-]*)$`)
)

// resolvePlaceholders is the Full-mode pass (FR-006..FR-010): it consumes @
// words that belong to placeholders in their command and replaces each
// {{name}} of a defined variable with a quoted value.
func resolvePlaceholders(line string, d *Defs, opt Options, res *Result) (string, error) {
	cmds := Tokenize(line)

	// placeholders per command; skip the whole pass if there are none
	type occ struct {
		word Word
		vars []string
	}
	perCmd := make([][]occ, len(cmds))
	found := false
	for i, c := range cmds {
		for _, w := range c.Words {
			if !w.Plain {
				continue
			}
			var vs []string
			for _, n := range model.PlaceholderNames(w.Text) {
				if _, ok := d.Vars[n]; ok {
					vs = append(vs, n)
				}
			}
			if len(vs) > 0 {
				perCmd[i] = append(perCmd[i], occ{w, vs})
				found = true
			}
		}
	}
	if !found {
		return line, nil
	}

	var edits []edit
	consume := func(w Word) {
		start := w.Start
		for start > 0 && (line[start-1] == ' ' || line[start-1] == '\t') {
			start--
		}
		edits = append(edits, edit{start, w.End, ""})
	}

	// @@env applies to the whole line
	lineEnv := ""
	for _, c := range cmds {
		for _, w := range c.Words {
			if m := atEnvRe.FindStringSubmatch(w.Text); m != nil && w.Plain {
				if _, ok := d.Envs[m[1]]; !ok {
					return "", fmt.Errorf("unknown environment %q in @@%s (see: ali env ls)", m[1], m[1])
				}
				lineEnv = m[1]
				consume(w)
			}
		}
	}

	termEnv, globalEnv := opt.TerminalEnv, opt.GlobalEnv
	if termEnv != "" {
		if _, ok := d.Envs[termEnv]; !ok {
			res.Warnings = append(res.Warnings, fmt.Sprintf("environment %q chosen in this terminal no longer exists; ignoring it", termEnv))
			termEnv = ""
		}
	}
	if globalEnv != "" {
		if _, ok := d.Envs[globalEnv]; !ok {
			res.Warnings = append(res.Warnings, fmt.Sprintf("global environment %q no longer exists; ignoring it", globalEnv))
			globalEnv = ""
		}
	}

	var missing []string
	for i, c := range cmds {
		if len(perCmd[i]) == 0 {
			continue
		}
		inCmd := map[string]bool{}
		for _, o := range perCmd[i] {
			for _, v := range o.vars {
				inCmd[v] = true
			}
		}
		explicit := map[string]string{}
		for _, w := range c.Words {
			if strings.HasPrefix(w.Text, "@@") {
				continue
			}
			if m := atSetRe.FindStringSubmatch(w.Text); m != nil && inCmd[m[1]] {
				raw, ok := unquote(w.Text[len(m[0]):])
				if !ok {
					continue // contains $ or backticks: never evaluated, left as typed
				}
				v := d.Vars[m[1]]
				if val, isKey := v.Lookup(raw); isKey {
					raw = val
				}
				explicit[m[1]] = raw
				consume(w)
				continue
			}
			if m := atKeyRe.FindStringSubmatch(w.Text); m != nil && w.Plain {
				var owners []string
				for v := range inCmd {
					if _, ok := d.Vars[v].Lookup(m[1]); ok {
						owners = append(owners, v)
					}
				}
				sort.Strings(owners)
				switch len(owners) {
				case 0:
					continue // not ours: leave the word alone
				case 1:
					explicit[owners[0]], _ = d.Vars[owners[0]].Lookup(m[1])
					consume(w)
				default:
					forms := make([]string, len(owners))
					for j, o := range owners {
						forms[j] = "@" + o + "=" + m[1]
					}
					return "", fmt.Errorf("@%s is ambiguous here; write %s", m[1], strings.Join(forms, " or "))
				}
			}
		}

		values := map[string]string{}
		for v := range inCmd {
			val, from, ok := resolveVar(d, v, explicit, lineEnv, termEnv, globalEnv)
			if !ok {
				missing = append(missing, v)
				continue
			}
			values[v] = val
			if from != "" && d.Envs[from].Protected {
				res.Protected, res.Env = true, from
			}
		}
		for _, o := range perCmd[i] {
			text := model.PlaceholderRe.ReplaceAllStringFunc(o.word.Text, func(ph string) string {
				name := ph[2 : len(ph)-2]
				if val, ok := values[name]; ok {
					return QuoteValue(val)
				}
				return ph
			})
			edits = append(edits, edit{o.word.Start, o.word.End, text})
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return "", fmt.Errorf("no value for %s: add @value or @%s=value, give the variable a default, or choose an environment (ali env use)",
			"{{"+strings.Join(missing, "}}, {{")+"}}", missing[0])
	}
	return apply(line, edits), nil
}

// resolveVar applies FR-009's precedence and reports which environment, if
// any, supplied the value.
func resolveVar(d *Defs, name string, explicit map[string]string, envs ...string) (val, fromEnv string, ok bool) {
	if v, ok := explicit[name]; ok {
		return v, "", true
	}
	variable := d.Vars[name]
	for _, e := range envs {
		if e == "" {
			continue
		}
		if set, ok := d.Envs[e].Set[name]; ok {
			if v, isKey := variable.Lookup(set); isKey {
				return v, e, true
			}
			return set, e, true
		}
	}
	if variable.Default != "" {
		v, _ := variable.Lookup(variable.Default)
		return v, "", true
	}
	return "", "", false
}

// unquote removes shell quoting from an @var= value. It refuses (ok=false)
// anything the shell would expand ($, backticks), because ali never evaluates
// shell code.
func unquote(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); {
		switch c := s[i]; c {
		case '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return "", false
			}
			b.WriteString(s[i+1 : i+1+j])
			i += j + 2
		case '"':
			i++
			for {
				if i >= len(s) {
					return "", false
				}
				if s[i] == '"' {
					i++
					break
				}
				if s[i] == '$' || s[i] == '`' {
					return "", false
				}
				if s[i] == '\\' && i+1 < len(s) && strings.IndexByte("\"\\$`", s[i+1]) >= 0 {
					i++
				}
				b.WriteByte(s[i])
				i++
			}
		case '\\':
			if i+1 >= len(s) {
				return "", false
			}
			b.WriteByte(s[i+1])
			i += 2
		case '$', '`':
			return "", false
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String(), true
}
