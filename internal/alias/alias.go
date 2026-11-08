// Package alias defines the Alias type and the validation rules for its parts.
package alias

import (
	"fmt"
	"regexp"
	"strings"
)

// Alias is one shell alias: a name that expands to a command, in a group.
type Alias struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Group   string `json:"group"`
}

var (
	nameRe  = regexp.MustCompile(`^[A-Za-z0-9_.][A-Za-z0-9_.@+:-]*$`)
	groupRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
)

// reservedGroups are top-level keys in the store file and cannot be groups.
var reservedGroups = map[string]bool{"version": true, "_pinned": true}

// reservedNames are words the generated shell code and the .zshrc wrapper rely
// on. Aliasing them could hijack that code (see ADR 0021).
var reservedNames = map[string]bool{
	"ali": true, "alias": true, "unalias": true, "builtin": true, "command": true,
	"source": true, "local": true, "return": true, "unset": true,
	"for": true, "do": true, "done": true,
}

// ValidateName reports why name is not a usable alias name.
func ValidateName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("alias name must not be empty")
	case len(name) > 64:
		return fmt.Errorf("alias name %q is longer than 64 characters", name)
	case reservedNames[name]:
		return fmt.Errorf("alias name %q is reserved: ali's shell integration relies on it", name)
	case name == ".":
		return fmt.Errorf("alias name %q is not allowed", name)
	case !nameRe.MatchString(name):
		return fmt.Errorf("invalid alias name %q: use letters, digits and _ . @ + : - (not starting with -)", name)
	}
	return nil
}

// ValidateGroup reports why group is not a usable group name.
func ValidateGroup(group string) error {
	if !groupRe.MatchString(group) || reservedGroups[group] {
		return fmt.Errorf("invalid group %q: use lowercase letters, digits, - and _ (not a reserved name)", group)
	}
	return nil
}

// ValidateCommand rejects empty commands and NUL bytes. Any other content is
// stored verbatim.
func ValidateCommand(cmd string) error {
	if strings.TrimSpace(cmd) == "" {
		return fmt.Errorf("command must not be empty")
	}
	if strings.ContainsRune(cmd, 0) {
		return fmt.Errorf("command must not contain a NUL byte")
	}
	return nil
}
