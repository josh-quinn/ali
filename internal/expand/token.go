// Package expand turns a command line typed with ali sub-aliases, placeholders
// and @ values into the command to run (ADR 0035). It does no I/O: the CLI
// passes in the definitions and settings, and the zsh glue applies the result.
package expand

import "strings"

// Word is one shell word with its exact byte span in the line.
type Word struct {
	Start, End int
	Text       string
	// Plain means the word is literal text: no quotes, backslashes, $ or
	// backticks. Only plain words can be aliases, sub-aliases, @ words or
	// placeholders; everything else is copied through untouched.
	Plain bool
}

// Command is the words between command separators (; | & && || newline ( )).
type Command struct{ Words []Word }

// Tokenize splits line into commands and words. It understands quoting,
// backslashes, $(...), `...`, <(...) and redirections such as 2>&1 well enough
// to never split inside them; it does not need to be a full shell parser
// because it only decides which words ali may touch.
func Tokenize(line string) []Command {
	var cmds []Command
	var cur Command
	i, n := 0, len(line)
	flush := func() {
		if len(cur.Words) > 0 {
			cmds = append(cmds, cur)
		}
		cur = Command{}
	}
	for i < n {
		c := line[i]
		switch {
		case c == ' ' || c == '\t':
			i++
			continue
		case c == '\n' || c == ';' || c == '(' || c == ')':
			flush()
			i++
			continue
		case c == '|' || (c == '&' && !(i+1 < n && line[i+1] == '>')):
			flush()
			i++
			continue
		}
		start := i
		plain := true
		for i < n {
			c := line[i]
			if c == ' ' || c == '\t' || c == '\n' || c == ';' || c == '|' || c == ')' {
				break
			}
			if c == '(' {
				break
			}
			if c == '&' {
				// part of a redirection: 2>&1, >&2, &>file
				if i > start && (line[i-1] == '>' || line[i-1] == '<') || (i+1 < n && line[i+1] == '>') {
					i++
					continue
				}
				break
			}
			switch c {
			case '\'':
				plain = false
				j := strings.IndexByte(line[i+1:], '\'')
				if j < 0 {
					i = n
				} else {
					i += j + 2
				}
				continue
			case '"':
				plain = false
				i = skipDouble(line, i+1)
				continue
			case '\\':
				plain = false
				i += 2
				if i > n {
					i = n
				}
				continue
			case '`':
				plain = false
				i = skipBacktick(line, i+1)
				continue
			case '$':
				plain = false
				if i+1 < n && line[i+1] == '(' {
					i = skipParens(line, i+2)
					continue
				}
			case '<', '>':
				if i+1 < n && line[i+1] == '(' { // process substitution
					plain = false
					i = skipParens(line, i+2)
					continue
				}
			}
			i++
		}
		cur.Words = append(cur.Words, Word{Start: start, End: i, Text: line[start:i], Plain: plain})
	}
	flush()
	return cmds
}

func skipDouble(s string, i int) int {
	for i < len(s) {
		switch s[i] {
		case '\\':
			i += 2
			continue
		case '"':
			return i + 1
		case '`':
			i = skipBacktick(s, i+1)
			continue
		case '$':
			if i+1 < len(s) && s[i+1] == '(' {
				i = skipParens(s, i+2)
				continue
			}
		}
		i++
	}
	return len(s)
}

func skipBacktick(s string, i int) int {
	for i < len(s) {
		switch s[i] {
		case '\\':
			i += 2
			continue
		case '`':
			return i + 1
		}
		i++
	}
	return len(s)
}

// skipParens returns the index just after the ")" matching an already-opened "(".
func skipParens(s string, i int) int {
	depth := 1
	for i < len(s) {
		switch s[i] {
		case '\\':
			i += 2
			continue
		case '\'':
			j := strings.IndexByte(s[i+1:], '\'')
			if j < 0 {
				return len(s)
			}
			i += j + 2
			continue
		case '"':
			i = skipDouble(s, i+1)
			continue
		case '`':
			i = skipBacktick(s, i+1)
			continue
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
		i++
	}
	return len(s)
}
