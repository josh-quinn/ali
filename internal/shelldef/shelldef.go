// Package shelldef reads what a name means in the user's live shell, as reported
// by the ali() wrapper (ADR 0027). The binary cannot see shell aliases itself.
package shelldef

import (
	"os"
	"strings"
)

// EnvVar is set by the wrapper for a single `ali add` invocation.
const EnvVar = "ALI_SHELL_DEF"

// Kind of definition a name has in the shell.
type Kind int

// Kinds of shell definitions.
const (
	None     Kind = iota // free, unknown, or not run through the wrapper
	Alias                // a non-ali alias
	Function             // a shell function
)

// Def is a name's definition in the live shell.
type Def struct {
	Kind Kind
	Text string // the alias expansion when Kind is Alias
}

// Parse decodes the wrapper's report: "alias<TAB>text", "function", or empty.
// Anything else is treated as None, so a malformed report never blocks ali.
func Parse(s string) Def {
	switch {
	case s == "function":
		return Def{Kind: Function}
	case strings.HasPrefix(s, "alias\t"):
		return Def{Kind: Alias, Text: strings.TrimPrefix(s, "alias\t")}
	}
	return Def{}
}

// FromEnv reads the report from the environment.
func FromEnv() Def { return Parse(os.Getenv(EnvVar)) }
