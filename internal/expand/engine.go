package expand

import (
	"regexp"
	"sort"
	"strings"

	"github.com/josh-quinn/ali/internal/group"
	"github.com/josh-quinn/ali/internal/model"
)

// Defs are the user's definitions, as loaded from the store.
type Defs struct {
	Aliases map[string]string            // ali alias name -> command
	Subs    map[string]map[string]string // tool -> sub-alias -> expansion
	Shared  map[string]string            // shared sub-alias -> expansion
	Vars    map[string]model.Variable
	Envs    map[string]model.Environment
}

// Mode says how much of a line to expand.
type Mode int

const (
	// Partial is used while typing (Space): aliases and sub-aliases only.
	Partial Mode = iota
	// Full is used when the line runs (Enter or run-time mode): also
	// placeholders and @ values.
	Full
)

// Options control one expansion.
type Options struct {
	Mode        Mode
	TerminalEnv string // ALI_ENV
	GlobalEnv   string // contents of global-env
}

// Result is the expanded line.
type Result struct {
	Line      string
	Changed   bool
	Protected bool     // a value came from a protected environment
	Env       string   // environment used, if any
	Warnings  []string // non-fatal notices (e.g. unknown ALI_ENV)
}

var (
	assignRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
	wrappers = map[string]bool{"sudo": true, "env": true, "command": true, "time": true, "nohup": true, "exec": true, "noglob": true, "nocorrect": true}
	// flags of wrappers that take a separate argument (sudo -u root, env -u X)
	wrapperArgFlags = map[string]bool{"-u": true, "-g": true, "-h": true, "-p": true, "-C": true, "-D": true, "-R": true, "-T": true}
)

type cmdInfo struct {
	tool    string
	cmdWord int    // index of the command word in the Command
	alias   string // ali alias used as command word, if any
}

// toolOf finds the command word (after assignments and wrappers) and the tool
// it runs, expanding an ali alias first. ADR 0008 rules apply.
func (d *Defs) toolOf(c Command) cmdInfo {
	inWrapper, skipNext := false, false
	for i, w := range c.Words {
		switch {
		case skipNext:
			skipNext = false
			continue
		case !w.Plain:
			return cmdInfo{cmdWord: -1}
		case assignRe.MatchString(w.Text):
			continue
		case wrappers[w.Text]:
			inWrapper = true
			continue
		case inWrapper && strings.HasPrefix(w.Text, "-"):
			skipNext = wrapperArgFlags[w.Text]
			continue
		}
		if cmd, ok := d.Aliases[w.Text]; ok {
			return cmdInfo{tool: group.Infer(cmd), cmdWord: i, alias: w.Text}
		}
		return cmdInfo{tool: group.Infer(w.Text), cmdWord: i}
	}
	return cmdInfo{cmdWord: -1}
}

// knownTool reports whether ali manages tool: it has its own sub-aliases or an
// ali alias runs it. Shared sub-aliases apply only to known tools, so a shared
// "o" never rewrites "cp o x".
func (d *Defs) knownTool(tool string) bool {
	if len(d.Subs[tool]) > 0 {
		return true
	}
	for _, cmd := range d.Aliases {
		if group.Infer(cmd) == tool {
			return true
		}
	}
	return false
}

func (d *Defs) sub(tool, word string) (string, bool) {
	if exp, ok := d.Subs[tool][word]; ok {
		return exp, true
	}
	if d.knownTool(tool) {
		exp, ok := d.Shared[word]
		return exp, ok
	}
	return "", false
}

type edit struct {
	start, end int
	text       string
}

// Expand expands line. In Partial mode only aliases and sub-aliases are
// expanded; Full mode also resolves placeholders and @ values.
func Expand(line string, d *Defs, opt Options) (Result, error) {
	var edits []edit
	for _, c := range Tokenize(line) {
		info := d.toolOf(c)
		if info.cmdWord < 0 || info.tool == "" {
			continue
		}
		var subEdits []edit
		for _, w := range c.Words[info.cmdWord+1:] {
			if !w.Plain {
				continue
			}
			if w.Text == "--" {
				break
			}
			if exp, ok := d.sub(info.tool, w.Text); ok {
				subEdits = append(subEdits, edit{w.Start, w.End, exp})
			}
		}
		// A plain alias only expands on the line when something else in the
		// command is expanded; otherwise zsh expands it as usual (FR-017).
		if info.alias != "" && len(subEdits) > 0 {
			w := c.Words[info.cmdWord]
			edits = append(edits, edit{w.Start, w.End, d.Aliases[info.alias]})
		}
		edits = append(edits, subEdits...)
	}
	return Result{Line: apply(line, edits), Changed: len(edits) > 0}, nil
}

func apply(line string, edits []edit) string {
	sort.Slice(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	var b strings.Builder
	at := 0
	for _, e := range edits {
		b.WriteString(line[at:e.start])
		b.WriteString(e.text)
		at = e.end
	}
	b.WriteString(line[at:])
	return b.String()
}
