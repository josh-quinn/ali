// Command ali creates, finds and organises shell aliases, sub-aliases,
// placeholders and environments for zsh. See README.md and "ali --help".
package main

import (
	"fmt"
	"os"

	"github.com/josh-quinn/ali/internal/cli"
	"github.com/josh-quinn/ali/internal/store"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	dir, err := store.DefaultDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ali:", err)
		os.Exit(1)
	}
	root := cli.NewRoot(version, store.New(dir), os.Stdout, os.Stderr)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ali:", err)
		if cli.IsUsage(err) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
