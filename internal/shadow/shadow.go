// Package shadow detects alias names that would hide an existing command.
package shadow

import (
	"os"
	"path/filepath"
)

var builtins = map[string]bool{}

func init() {
	for _, n := range []string{
		"cd", "echo", "export", "source", "alias", "unalias", "type", "which", "pwd", "set", "unset",
		"read", "eval", "exec", "exit", "return", "test", "true", "false", "history", "jobs", "fg", "bg",
		"kill", "wait", "trap", "ulimit", "umask", "printf", "print", "local", "typeset", "declare",
		"if", "then", "else", "elif", "fi", "for", "while", "until", "do", "done", "case", "esac",
		"function", "select", "time", "command", "builtin", "hash", "rehash", "autoload", "setopt", "unsetopt",
	} {
		builtins[n] = true
	}
}

// Shadows returns a description of what name would hide ("shell builtin" or
// the executable's path), or "" if nothing.
func Shadows(name string) string {
	if builtins[name] {
		return "shell builtin or keyword"
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		p := filepath.Join(dir, name)
		// Looking names up on $PATH is the purpose here; name is a validated alias name.
		if st, err := os.Stat(p); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 { // #nosec G703 G304
			return p
		}
	}
	return ""
}
