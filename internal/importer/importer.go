// Package importer extracts simple `alias name=value` lines from a zsh rc file.
// It imports only what it can read with certainty and reports the rest.
package importer

import (
	"fmt"
	"strings"
)

// Entry is an alias definition found in the file.
type Entry struct {
	Name, Command string
	LineNo        int
	Line          string
}

// Skip is a line that looked like an alias but was not imported.
type Skip struct {
	Line, Reason string
	LineNo       int
}

// Result is the outcome of parsing a file.
type Result struct {
	Entries []Entry
	Skipped []Skip
}

// Parse scans src line by line. Multi-line values and expansions ($, `) in
// double quotes are skipped because importing them literally would change meaning.
func Parse(src string) Result {
	var r Result
	for i, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, "alias ") && !strings.HasPrefix(line, "alias\t") {
			continue
		}
		rest := strings.TrimSpace(line[len("alias"):])
		skip := func(reason string) { r.Skipped = append(r.Skipped, Skip{Line: line, Reason: reason, LineNo: i + 1}) }
		if strings.HasPrefix(rest, "-") {
			skip("alias flags (-g/-s) are not supported")
			continue
		}
		eq := strings.IndexByte(rest, '=')
		if eq <= 0 {
			skip("not a name=value definition")
			continue
		}
		name := rest[:eq]
		if strings.ContainsAny(name, " \t") {
			skip("name contains whitespace")
			continue
		}
		val, tail, err := word(rest[eq+1:])
		if err != nil {
			skip(err.Error())
			continue
		}
		if t := strings.TrimSpace(tail); t != "" && !strings.HasPrefix(t, "#") {
			skip("multiple definitions or trailing text on one line")
			continue
		}
		if tail != "" && !strings.HasPrefix(tail, " ") && !strings.HasPrefix(tail, "\t") {
			skip("trailing text after value")
			continue
		}
		if strings.TrimSpace(val) == "" {
			skip("empty command")
			continue
		}
		r.Entries = append(r.Entries, Entry{Name: name, Command: val, LineNo: i + 1, Line: line})
	}
	return r
}

// word reads one shell word (concatenated bare/single/double segments) and
// returns its value and the unread remainder.
func word(s string) (val, rest string, err error) {
	var b strings.Builder
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == ' ' || c == '\t':
			return b.String(), s[i:], nil
		case c == '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return "", "", fmt.Errorf("unterminated single quote (multi-line value?)")
			}
			b.WriteString(s[i+1 : i+1+j])
			i += j + 2
		case c == '"':
			i++
			closed := false
			for i < len(s) {
				d := s[i]
				if d == '\\' && i+1 < len(s) && strings.IndexByte(`"\$`+"`", s[i+1]) >= 0 {
					b.WriteByte(s[i+1])
					i += 2
					continue
				}
				if d == '$' || d == '`' {
					return "", "", fmt.Errorf("double-quoted value uses expansion ($ or backtick); convert it to single quotes first")
				}
				if d == '"' {
					closed = true
					i++
					break
				}
				b.WriteByte(d)
				i++
			}
			if !closed {
				return "", "", fmt.Errorf("unterminated double quote (multi-line value?)")
			}
		case c == '\\':
			if i+1 >= len(s) {
				return "", "", fmt.Errorf("line ends with a backslash")
			}
			b.WriteByte(s[i+1])
			i += 2
		case c == '$' || c == '`':
			return "", "", fmt.Errorf("unquoted expansion ($ or backtick) is not supported")
		case strings.IndexByte(";&|<>()", c) >= 0:
			return "", "", fmt.Errorf("unquoted %q ends the alias definition in zsh (quote it or write it by hand)", c)
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String(), "", nil
}
