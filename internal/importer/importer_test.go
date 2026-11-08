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
