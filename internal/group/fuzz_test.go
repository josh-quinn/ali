package group

import (
	"testing"

	"github.com/josh-quinn/ali/internal/alias"
)

// Infer always returns a valid group name, whatever the command.
func FuzzInfer(f *testing.F) {
	for _, s := range []string{"git pull", "sudo -E helm ls", "FOO=1 docker ps", "./run.sh", "", "!!", "'quoted tool' x"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, cmd string) {
		g := Infer(cmd)
		if err := alias.ValidateGroup(g); err != nil {
			t.Fatalf("Infer(%q) = %q, which is not a valid group: %v", cmd, g, err)
		}
	})
}
