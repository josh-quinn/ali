package expand

import (
	"strings"
	"testing"

	"github.com/josh-quinn/ali/internal/render"
)

// Tokenize must never panic, and every word must be the exact, ordered,
// non-overlapping slice of the line it reports.
func FuzzTokenize(f *testing.F) {
	for _, s := range []string{"k g p", "echo 'a; b' \"c | d\"", "echo $(k g | wc -l) x", "diff <(a) <(b)", "a\\ b 2>&1 &>log", "'unterminated", "(cd x; k g)", "@ns='a b' {{ns}}"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, line string) {
		end := 0
		for _, c := range Tokenize(line) {
			for _, w := range c.Words {
				if w.Start < end || w.End < w.Start || w.End > len(line) {
					t.Fatalf("bad span %d:%d (previous end %d) in %q", w.Start, w.End, end, line)
				}
				if line[w.Start:w.End] != w.Text || w.Text == "" {
					t.Fatalf("span %d:%d = %q, Text %q in %q", w.Start, w.End, line[w.Start:w.End], w.Text, line)
				}
				end = w.End
			}
		}
	})
}

// Quoting must round-trip any value: what ali inserts is exactly what the
// user gave (ADR 0013, 0036).
func FuzzQuoteRoundTrip(f *testing.F) {
	for _, s := range []string{"", "it's", "a b; rm -rf ~", "$(id)", "`x`", "\\", "'''", "a\nb", "=ls", "é ✓"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if strings.ContainsRune(s, 0) {
			t.Skip("NUL cannot be stored (rejected by validation)")
		}
		for _, q := range []string{render.Quote(s), QuoteValue(s)} {
			if s == "" && q == "" {
				continue
			}
			got, ok := unquote(q)
			if !ok || got != s {
				t.Fatalf("value %q quoted as %q unquotes to %q (ok=%v)", s, q, got, ok)
			}
		}
	})
}

// Expansion must never panic, and must never change text inside quotes.
func FuzzExpand(f *testing.F) {
	for _, s := range []string{"k g p n @sys", "k g p c", "kgp @ns='a b'", "sudo -u root k g", "k g | h g; kubectl d p", "k exec x -- g", "k g p n @@dev", "npm i @types/node"} {
		f.Add(s, true)
	}
	f.Fuzz(func(t *testing.T, line string, fullMode bool) {
		mode := Partial
		if fullMode {
			mode = Full
		}
		res, err := Expand(line, vdefs(), Options{Mode: mode, TerminalEnv: "dev"})
		if err != nil {
			return // a refused line is fine; it is reported, not run
		}
		if !res.Changed && res.Line != line {
			t.Fatalf("unchanged result differs: %q -> %q", line, res.Line)
		}
		// every single-quoted word of the input survives byte for byte
		for _, c := range Tokenize(line) {
			for _, w := range c.Words {
				if strings.HasPrefix(w.Text, "'") && strings.HasSuffix(w.Text, "'") && len(w.Text) > 1 && !strings.Contains(w.Text[1:len(w.Text)-1], "'") {
					if !strings.Contains(res.Line, w.Text) {
						t.Fatalf("quoted word %q lost: %q -> %q", w.Text, line, res.Line)
					}
				}
			}
		}
	})
}
