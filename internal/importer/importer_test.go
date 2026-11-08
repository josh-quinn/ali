package importer

import "testing"

func TestParse(t *testing.T) {
	src := `# my zshrc
export FOO=bar
alias k=kubectl
alias gs='git status'
alias   ll="ls -la"
alias q='echo '\''hi'\'''
alias esc=ls\ -a
alias c='cd ..' # go up
alias -g G='| grep'
alias a=1 b=2
alias dq="echo $HOME"
alias bt="echo ` + "`date`" + `"
alias plain
alias bad name=x
  alias indented='ls'
alias empty=
alias ml='echo a
b'
`
	r := Parse(src)

	want := map[string]string{
		"k":        "kubectl",
		"gs":       "git status",
		"ll":       "ls -la",
		"q":        "echo 'hi'",
		"esc":      "ls -a",
		"c":        "cd ..",
		"indented": "ls",
	}
	if len(r.Entries) != len(want) {
		t.Errorf("got %d entries %+v", len(r.Entries), r.Entries)
	}
	for _, e := range r.Entries {
		if want[e.Name] != e.Command {
			t.Errorf("%s = %q, want %q", e.Name, e.Command, want[e.Name])
		}
	}
	skipped := map[string]bool{}
	for _, s := range r.Skipped {
		skipped[s.Line] = true
		if s.Reason == "" {
			t.Errorf("skipped %q without reason", s.Line)
		}
	}
	for _, l := range []string{"alias -g G='| grep'", "alias a=1 b=2", `alias dq="echo $HOME"`, "alias plain", "alias bad name=x", "alias empty="} {
		if !skipped[l] {
			t.Errorf("expected %q to be skipped; skipped=%v", l, r.Skipped)
		}
	}
}

func TestLineNumbers(t *testing.T) {
	r := Parse("x\nalias a=b\n")
	if len(r.Entries) != 1 || r.Entries[0].LineNo != 2 {
		t.Errorf("%+v", r.Entries)
	}
}

// In `alias a=b;c` zsh defines alias a=b and then *runs* c. Importing "b;c" as the
// alias value would misrepresent the file.
func TestUnquotedShellMetacharactersAreNotImported(t *testing.T) {
	lines := []string{"alias a=b;c", "alias d=e&&f", "alias p=a|b", "alias r=a>f", "alias s=a<f", "alias t=(x)", "alias u=a&"}
	for _, l := range lines {
		r := Parse(l + "\n")
		if len(r.Entries) != 0 || len(r.Skipped) != 1 {
			t.Errorf("%q: entries=%v skipped=%v", l, r.Entries, r.Skipped)
		}
	}
	// the same characters inside quotes are ordinary text
	r := Parse("alias ok='a;b && c | d > e'\nalias ok2=\"x;y\"\nalias ok3=a\\;b\n")
	if len(r.Entries) != 3 {
		t.Fatalf("quoted/escaped metacharacters should import: %+v %+v", r.Entries, r.Skipped)
	}
	if r.Entries[0].Command != "a;b && c | d > e" || r.Entries[2].Command != "a;b" {
		t.Errorf("wrong values: %+v", r.Entries)
	}
}
