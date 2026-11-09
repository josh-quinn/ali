// Command ali creates, finds and organises shell aliases, sub-aliases,
// placeholders and environments for zsh. See README.md and "ali --help".
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/josh-quinn/ali/internal/cli"
	"github.com/josh-quinn/ali/internal/store"
)

// version is set at build time via -ldflags "-X main.version=..." (releases).
var version = "dev"

// resolveVersion prefers the stamped version and otherwise uses the module
// version Go records for "go install module@version" builds.
func resolveVersion(stamped string, info *debug.BuildInfo) string {
	if stamped != "dev" || info == nil {
		return stamped
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return strings.TrimPrefix(v, "v")
	}
	return stamped
}

func main() {
	dir, err := store.DefaultDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ali:", err)
		os.Exit(1)
	}
	info, _ := debug.ReadBuildInfo()
	root := cli.NewRoot(resolveVersion(version, info), store.New(dir), os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ali:", err)
		if cli.IsUsage(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
