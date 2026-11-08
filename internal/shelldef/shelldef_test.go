package shelldef

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Def
	}{
		{"", Def{}},
		{"alias\tgit push", Def{Kind: Alias, Text: "git push"}},
		{"alias\techo 'a\tb'", Def{Kind: Alias, Text: "echo 'a\tb'"}},
		{"alias\t", Def{Kind: Alias, Text: ""}},
		{"function", Def{Kind: Function}},
		{"garbage", Def{}},
		{"alias", Def{}},
		{"function\textra", Def{}},
	}
	for _, c := range cases {
		if got := Parse(c.in); got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestFromEnv(t *testing.T) {
	t.Setenv(EnvVar, "function")
	if FromEnv().Kind != Function {
		t.Error("FromEnv did not read " + EnvVar)
	}
}
