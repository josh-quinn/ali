// Package group infers an alias group from the command it runs.
package group

import (
	"path/filepath"
	"regexp"
	"strings"
)

const fallback = "misc"

var (
	assignRe  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)
	invalidRe = regexp.MustCompile(`[^a-z0-9_-]+`)
	wrappers  = map[string]bool{"sudo": true, "env": true, "command": true, "time": true, "nohup": true, "exec": true}
)

// Infer returns the group for a command: the base name of the first word,
// skipping wrappers (sudo, env, ...), their flags, and VAR=value assignments.
func Infer(command string) string {
	skipFlags := false
	for _, tok := range strings.Fields(command) {
		tok = strings.Trim(tok, `'"`)
		switch {
		case tok == "":
			continue
		case assignRe.MatchString(tok):
			continue
		case skipFlags && strings.HasPrefix(tok, "-"):
			continue
		case wrappers[tok]:
			skipFlags = true
			continue
		}
		return clean(filepath.Base(tok))
	}
	return fallback
}

func clean(s string) string {
	s = invalidRe.ReplaceAllString(strings.ToLower(s), "-")
	s = strings.Trim(s, "-_")
	if s == "" {
		return fallback
	}
	return s
}
