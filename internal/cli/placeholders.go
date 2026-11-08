package cli

import (
	"fmt"

	"github.com/josh-quinn/ali/internal/expand"
	"github.com/josh-quinn/ali/internal/model"
	"github.com/josh-quinn/ali/internal/store"
)

// checkPlaceholders validates {{name}} placeholders in a command or sub-alias
// expansion being saved (FR-006). A placeholder of a defined variable inside
// quotes is an error, because no value can be inserted there safely; a
// {{name}} that is not a variable is literal text and only gets a note.
func (a *app) checkPlaceholders(d *store.Data, text string) error {
	for _, c := range expand.Tokenize(text) {
		for _, w := range c.Words {
			for _, n := range model.PlaceholderNames(w.Text) {
				_, defined := d.Var(n)
				switch {
				case defined && !w.Plain:
					return fmt.Errorf("placeholder {{%s}} is inside quotes or next to $ or \\, where ali cannot insert a value safely; put it outside quotes, e.g. -n {{%s}}", n, n)
				case !defined && w.Plain:
					fmt.Fprintf(a.err, "note: {{%s}} is not a variable, so it is kept as literal text (create it with: ali var add %s KEY=VALUE)\n", n, n)
				}
			}
		}
	}
	return nil
}
