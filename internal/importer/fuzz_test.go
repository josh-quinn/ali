package importer

import (
	"strings"
	"testing"
)

// Parse must never panic; every entry has a name, a non-empty command, and a
// line number inside the input.
func FuzzParse(f *testing.F) {
	for _, s := range []string{"alias k=kubectl\n", "alias gs='git status'\nalias q='echo '\\''hi'\\'''\n", "alias a=b;c\n", "alias dq=\"echo $HOME\"\n", "alias ml='a\nb'\n"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		r := Parse(src)
		lines := strings.Count(src, "\n") + 1
		for _, e := range r.Entries {
			if e.Name == "" || strings.TrimSpace(e.Command) == "" || e.LineNo < 1 || e.LineNo > lines {
				t.Fatalf("bad entry %+v from %q", e, src)
			}
		}
		for _, s := range r.Skipped {
			if s.Reason == "" || s.LineNo < 1 || s.LineNo > lines {
				t.Fatalf("bad skip %+v from %q", s, src)
			}
		}
	})
}
