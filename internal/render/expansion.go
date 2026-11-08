package render

import (
	"sort"
	"strings"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/group"
	"github.com/josh-quinn/ali/internal/model"
)

// Input is everything aliases.zsh is generated from.
type Input struct {
	Aliases   []alias.Alias
	Subs      map[string]map[string]string // tool -> sub-alias -> expansion
	Shared    map[string]string
	Vars      []string // defined variable names
	Envs      []string // environment names
	Protected []string // protected environment names
	PromptOff bool     // settings: prompt = off
}

// Render produces aliases.zsh: the aliases (Zsh), plus, only when sub-aliases
// or placeholders exist, the tables and line-editor widgets for visible
// expansion (ADRs 0030, 0035). Users with plain aliases get exactly Zsh's output.
func Render(in Input) string {
	plain, funcs := Split(in)
	out := Zsh(plain) + functionsCode(funcs) + envCode(in)
	trigger, subw := gateTables(in)
	if len(trigger) == 0 {
		return out
	}
	var b strings.Builder
	b.WriteString(out)
	b.WriteString("\\typeset -gA _ali_trigger _ali_subw\n")
	b.WriteString("_ali_trigger=(" + assocOnes(trigger) + ")\n")
	b.WriteString("_ali_subw=(" + assocOnes(subw) + ")\n")
	b.WriteString(widgets)
	return b.String()
}

// Split separates aliases rendered as plain zsh aliases from those rendered as
// functions. An alias containing a placeholder becomes a function that asks ali
// to expand it, so it can never run with a literal {{name}} (for example when
// the line editor is not involved).
func Split(in Input) (plain, funcs []alias.Alias) {
	vars := map[string]bool{}
	for _, v := range in.Vars {
		vars[v] = true
	}
	for _, a := range in.Aliases {
		isFunc := false
		for _, p := range model.PlaceholderNames(a.Command) {
			if vars[p] {
				isFunc = true
			}
		}
		if isFunc {
			funcs = append(funcs, a)
		} else {
			plain = append(plain, a)
		}
	}
	return plain, funcs
}

// functionsCode defines the function-rendered aliases and removes functions
// ali defined earlier in this session that are no longer needed.
func functionsCode(funcs []alias.Alias) string {
	names := make([]string, len(funcs))
	for i, a := range funcs {
		names[i] = Quote(a.Name)
	}
	var b strings.Builder
	b.WriteString("\\typeset -ga _ali_funcs\n")
	b.WriteString("_ali_fold=($_ali_funcs)\n")
	b.WriteString("_ali_funcs=(" + strings.Join(names, " ") + ")\n")
	b.WriteString("for _n in ${_ali_fold:|_ali_funcs}; do \\unfunction -- \"$_n\" 2>/dev/null; done\n")
	b.WriteString("\\unset _n _ali_fold\n")
	for _, a := range funcs {
		body := `\local _c; _c="$(\command ali expand --run -- ` + Quote(a.Name) + ` "$@")" || \return $?; \eval "$_c"`
		// The key is not quoted: quotes in a subscript would become part of the
		// function's name. Alias names are validated to safe characters.
		b.WriteString("functions[" + a.Name + "]=" + Quote(body) + "\n")
	}
	return b.String()
}

// envCode tells the prompt hook which environments are protected and whether
// the indicator is off. Nothing is emitted for users without environments.
func envCode(in Input) string {
	if len(in.Envs) == 0 && !in.PromptOff {
		return ""
	}
	prot := map[string]bool{}
	for _, p := range in.Protected {
		prot[p] = true
	}
	s := "\\typeset -gA _ali_protected\n_ali_protected=(" + assocOnes(prot) + ")\n"
	if in.PromptOff {
		return s + "_ali_prompt_off=1\n"
	}
	return s + "\\unset _ali_prompt_off\n"
}

// gateTables computes, for the zsh side, which first words can start a line ali
// may expand (trigger) and which words finishing on Space can expand (subw),
// so ordinary typing never starts a process (research R2).
func gateTables(in Input) (trigger, subw map[string]bool) {
	trigger, subw = map[string]bool{}, map[string]bool{}
	vars := map[string]bool{}
	for _, v := range in.Vars {
		vars[v] = true
	}
	known := map[string]bool{} // tools ali knows (engine.knownTool)
	for tool, subs := range in.Subs {
		if len(subs) == 0 {
			continue
		}
		known[tool] = true
		for n := range subs {
			subw[n] = true
		}
	}
	for n := range in.Shared {
		subw[n] = true
	}
	if len(in.Shared) > 0 {
		for _, a := range in.Aliases {
			known[group.Infer(a.Command)] = true
		}
	}
	for tool := range known {
		trigger[tool] = true
	}
	for _, a := range in.Aliases {
		if known[group.Infer(a.Command)] {
			trigger[a.Name] = true
		}
		for _, p := range model.PlaceholderNames(a.Command) {
			if vars[p] {
				trigger[a.Name] = true
				subw[a.Name] = true
			}
		}
	}
	return trigger, subw
}

func assocOnes(set map[string]bool) string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = Quote(k) + " 1"
	}
	return strings.Join(parts, " ")
}

// widgets is the line-editor glue (research R3, R4). Builtins are escaped
// (ADR 0021). Previously bound widgets are captured once, so re-sourcing this
// file after every ali command never makes ali wrap itself.
const widgets = `# ali: show sub-alias and placeholder expansion in the command line (ADR 0030)
_ali_gate() {
  \local _w
  for _w in ${(z)1}; do (( ${+_ali_trigger[$_w]} )) && \return 0; done
  \return 1
}
_ali_space() {
  # No PENDING check: fast typists type ahead while ali runs, and pastes do not
  # reach this widget when bracketed paste is on (the zsh default since 5.1).
  if (( ${+_ali_subw[${${(z)LBUFFER}[-1]}]} )) && _ali_gate "$LBUFFER"; then
    \local -a _r
    _r=("${(@0)$(\command ali expand --zle space -- "$LBUFFER" 2>/dev/null)}")
    [[ $_r[1] == ok ]] && LBUFFER=$_r[2]
  fi
  \zle _ali_orig_space
}
_ali_accept() {
  if [[ -n $_ali_pending && $BUFFER == "$_ali_pending" ]]; then
    _ali_pending=
    \zle _ali_orig_accept
    \return
  fi
  _ali_pending=
  if _ali_gate "$BUFFER"; then
    \local -a _r
    _r=("${(@0)$(\command ali expand --zle accept -- "$BUFFER" 2>/dev/null)}")
    if [[ $_r[1] == ok ]]; then
      BUFFER=$_r[2]
      CURSOR=$#BUFFER
    elif [[ $_r[1] == confirm ]]; then
      BUFFER=$_r[2]
      CURSOR=$#BUFFER
      _ali_pending=$BUFFER
      \zle -M "$_r[3]"
      \return
    elif [[ $_r[1] == error ]]; then
      \zle -M "$_r[3]"
      \return
    fi
    [[ -n $_r[3] ]] && \zle -M "$_r[3]"
  fi
  \zle _ali_orig_accept
}
\zmodload -i zsh/zleparameter 2>/dev/null
if (( ! ${+widgets[_ali_orig_accept]} )); then
  \zle -A accept-line _ali_orig_accept
  \zle -N accept-line _ali_accept
  \zle -N _ali_accept
  _ali_sp=${${(z)"$(\bindkey -M main ' ')"}[2]}
  [[ -z $_ali_sp || $_ali_sp == _ali_space ]] && _ali_sp=self-insert
  \zle -A $_ali_sp _ali_orig_space
  \zle -N _ali_space
  \bindkey -M main ' ' _ali_space
  \bindkey -M viins ' ' _ali_space 2>/dev/null
  \unset _ali_sp
fi
`
