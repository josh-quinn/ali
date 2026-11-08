package cli

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/josh-quinn/ali/internal/alias"
	"github.com/josh-quinn/ali/internal/store"
)

func TestListAndSearchStayFastAt1000Aliases(t *testing.T) {
	e := newEnv(t)
	err := e.st.Update(func(d *store.Data) error {
		for g := 0; g < 50; g++ {
			for i := 0; i < 20; i++ {
				d.Set(alias.Alias{Name: fmt.Sprintf("a%dx%d", g, i), Command: fmt.Sprintf("tool%d run %d", g, i), Group: fmt.Sprintf("tool%d", g)}, false)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"ls"}, {"search", "tool49"}, {"groups"}} {
		start := time.Now()
		e.must(args...)
		if took := time.Since(start); took > 50*time.Millisecond {
			t.Errorf("ali %v took %v, want < 50ms (SC-003)", args, took)
		}
	}
	if out := e.must("groups"); strings.Count(out, "\n") != 50 {
		t.Errorf("expected 50 groups, got:\n%s", out)
	}
}
