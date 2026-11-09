package store

import (
	"testing"
)

// Parsing arbitrary bytes must never panic, and anything accepted must survive
// a save and reload unchanged.
func FuzzParse(f *testing.F) {
	for _, s := range []string{"version = 1\n[git]\ngp = \"git pull\"\n", "version = 2\n[_subs.kubectl]\ng = \"get\"\n[_vars.ns.values]\none = \"x\"\n", "[oops\n", "version = 99\n"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		d, err := parse("aliases.toml", raw)
		if err != nil {
			return
		}
		out, err := d.marshal()
		if err != nil {
			t.Fatalf("accepted store cannot be saved: %v", err)
		}
		d2, err := parse("aliases.toml", out)
		if err != nil {
			t.Fatalf("saved store cannot be reloaded: %v\n%s", err, out)
		}
		a, b := d.All(), d2.All()
		if len(a) != len(b) {
			t.Fatalf("aliases changed on round trip: %v vs %v", a, b)
		}
		for i := range a {
			if a[i] != b[i] {
				t.Fatalf("alias changed on round trip: %+v vs %+v", a[i], b[i])
			}
		}
	})
}
