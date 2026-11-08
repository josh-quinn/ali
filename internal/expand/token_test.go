package expand

import (
	"reflect"
	"testing"
)

func texts(cmds []Command) [][]string {
	var out [][]string
	for _, c := range cmds {
		var ws []string
		for _, w := range c.Words {
			ws = append(ws, w.Text)
		}
		out = append(out, ws)
	}
	return out
}

func TestTokenizeSplitsCommands(t *testing.T) {
	cases := map[string][][]string{
		"k g p":                      {{"k", "g", "p"}},
		"  k   g\tp  ":               {{"k", "g", "p"}},
		"k g p; h g":                 {{"k", "g", "p"}, {"h", "g"}},
		"k g p | grep x && h g || y": {{"k", "g", "p"}, {"grep", "x"}, {"h", "g"}, {"y"}},
		"k g p &":                    {{"k", "g", "p"}},
		"k g p 2>&1 >/dev/null":      {{"k", "g", "p", "2>&1", ">/dev/null"}},
		"k g p &>log":                {{"k", "g", "p", "&>log"}},
		"echo 'a; b' \"c | d\"":      {{"echo", "'a; b'", "\"c | d\""}},
		"echo $(k g | wc -l) x":      {{"echo", "$(k g | wc -l)", "x"}},
		"echo `k g; p` x":            {{"echo", "`k g; p`", "x"}},
		"diff <(k g) <(h g)":         {{"diff", "<(k g)", "<(h g)"}},
		"a\nb":                       {{"a"}, {"b"}},
		"echo a\\ b c":               {{"echo", "a\\ b", "c"}},
		"(cd x; k g)":                {{"cd", "x"}, {"k", "g"}},
		"":                           nil,
		"echo 'unterminated":         {{"echo", "'unterminated"}},
		"--namespace={{ns}} @sys":    {{"--namespace={{ns}}", "@sys"}},
	}
	for in, want := range cases {
		if got := texts(Tokenize(in)); !reflect.DeepEqual(got, want) {
			t.Errorf("Tokenize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPlainWords(t *testing.T) {
	plain := map[string]bool{
		"g": true, "{{ns}}": true, "@sys": true, "--": true, "x=y": true, "*.go": true,
		`\p`: false, "'g'": false, `"p"`: false, "$x": false, "a`b`": false, "$(x)": false, "'unterminated": false,
	}
	for w, want := range plain {
		cmds := Tokenize("cmd " + w)
		if len(cmds) != 1 || len(cmds[0].Words) != 2 {
			t.Fatalf("%q tokenized as %q", w, texts(cmds))
		}
		if got := cmds[0].Words[1].Plain; got != want {
			t.Errorf("%q plain = %v, want %v", w, got, want)
		}
	}
}

func TestSpansPointIntoTheLine(t *testing.T) {
	line := "  k 'g x'  p;h"
	for _, c := range Tokenize(line) {
		for _, w := range c.Words {
			if line[w.Start:w.End] != w.Text {
				t.Errorf("span %d:%d = %q, text %q", w.Start, w.End, line[w.Start:w.End], w.Text)
			}
		}
	}
}
