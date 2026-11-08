package expand

import (
	"regexp"

	"github.com/josh-quinn/ali/internal/render"
)

var safeValueRe = regexp.MustCompile(`^[A-Za-z0-9_./:=,%+@-]+$`)

// QuoteValue makes a substituted value exactly one literal shell word (ADR
// 0036): safe values stay readable, anything else is single-quoted. A leading
// "=" is quoted because zsh would expand "=cmd" to the command's path.
func QuoteValue(v string) string {
	if safeValueRe.MatchString(v) && v[0] != '=' {
		return v
	}
	return render.Quote(v)
}
