package expand

import (
	"reflect"
	"testing"
)

func TestCompleteValuesAndSubAliases(t *testing.T) {
	d := vdefs()
	cases := []struct {
		lbuf string
		want []string
	}{
		{"k g p n @", []string{"@dev", "@one", "@sys"}},
		{"k g p n @s", []string{"@sys"}},
		{"kgp @", []string{"@dev", "@one", "@sys"}},                                 // placeholder inside an alias
		{"k g p n c @", []string{"@ctx=", "@dev", "@ns=", "@one", "@prod", "@sys"}}, // several variables: also @var= forms
		{"k g p n @@", []string{"@@dev", "@@prod"}},
		{"k g p @", nil}, // no placeholder in the line
		{"k ", []string{"c", "d", "g", "n", "o", "p", "w"}}, // kubectl's own plus shared
		{"k g p", []string{"p"}},                            // completing a partial word
		{"h ", []string{"g", "o", "w"}},
		{"ls ", nil}, // unknown tool
		{"echo @", nil},
	}
	for _, c := range cases {
		if got := Complete(c.lbuf, d); !reflect.DeepEqual(got, c.want) {
			t.Errorf("Complete(%q) = %q, want %q", c.lbuf, got, c.want)
		}
	}
}
