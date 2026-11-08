// Package model defines the expansion objects of feature 003 (sub-aliases,
// variables and environments) and their validation rules (data-model.md).
package model

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	subRe  = regexp.MustCompile(`^[A-Za-z0-9_.][A-Za-z0-9_.+:-]*$`)
	toolRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	varRe  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	keyRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	envRe  = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	// PlaceholderRe matches {{name}}; only names of defined variables count.
	PlaceholderRe = regexp.MustCompile(`\{\{([a-z][a-z0-9_]*)\}\}`)
)

func check(re *regexp.Regexp, what, s, hint string) error {
	if !re.MatchString(s) {
		return fmt.Errorf("invalid %s %q: %s", what, s, hint)
	}
	return nil
}

// ValidateSubName checks a sub-alias name (it never starts with @ or \, which
// have their own meaning on the command line).
func ValidateSubName(s string) error {
	return check(subRe, "sub-alias name", s, "use letters, digits and _ . + : - (not starting with -)")
}

// ValidateTool checks a tool name (the command a line runs, e.g. kubectl).
func ValidateTool(s string) error {
	return check(toolRe, "tool", s, "use the command's lowercase name, e.g. kubectl")
}

// ValidateVarName checks a variable name; "env" is reserved for @@env syntax.
func ValidateVarName(s string) error {
	if s == "env" {
		return fmt.Errorf("variable name %q is reserved", s)
	}
	return check(varRe, "variable name", s, "use lowercase letters, digits and _, starting with a letter")
}

// ValidateKey checks the name of a variable's value (what follows @).
func ValidateKey(s string) error {
	return check(keyRe, "value name", s, "use lowercase letters, digits, - and _")
}

// ValidateEnvName checks an environment name.
func ValidateEnvName(s string) error {
	return check(envRe, "environment name", s, "use lowercase letters, digits, - and _, starting with a letter")
}

// ValidateValue checks a variable value. Values are data and may contain any
// text (they are inserted as single quoted words, ADR 0036) except NUL.
func ValidateValue(s string) error {
	switch {
	case s == "":
		return fmt.Errorf("value must not be empty")
	case strings.ContainsRune(s, 0):
		return fmt.Errorf("value must not contain a NUL byte")
	}
	return nil
}

// PlaceholderNames lists the distinct {{name}} placeholders in text, in order.
func PlaceholderNames(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range PlaceholderRe.FindAllStringSubmatch(text, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			out = append(out, m[1])
		}
	}
	return out
}

// Variable is a named slot with named values and an optional default key.
type Variable struct {
	Values  map[string]string
	Default string
}

// Lookup returns the value stored under key.
func (v Variable) Lookup(key string) (string, bool) {
	val, ok := v.Values[key]
	return val, ok
}

// Validate checks keys, values and the default.
func (v Variable) Validate() error {
	for k, val := range v.Values {
		if err := ValidateKey(k); err != nil {
			return err
		}
		if err := ValidateValue(val); err != nil {
			return fmt.Errorf("value %s: %w", k, err)
		}
	}
	if v.Default != "" {
		if _, ok := v.Values[v.Default]; !ok {
			return fmt.Errorf("default %q is not one of the values", v.Default)
		}
	}
	return nil
}

// Environment assigns values (a key of the variable, or literal text) to variables.
type Environment struct {
	Protected bool
	Set       map[string]string
}
