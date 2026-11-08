// Command gendocs writes ali's man pages from the cobra command tree, so they
// always match --help (ADR 0038). It is a build tool and is not part of ali.
//
//	go run ./internal/tools/gendocs -out man -version v1.0.0
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra/doc"

	"github.com/josh-quinn/ali/internal/cli"
	"github.com/josh-quinn/ali/internal/store"
)

func main() {
	out := flag.String("out", "man", "directory to write man pages into")
	version := flag.String("version", "dev", "version printed in the pages")
	flag.Parse()
	if err := generate(*out, *version); err != nil {
		fmt.Fprintln(os.Stderr, "gendocs:", err)
		os.Exit(1)
	}
}

func generate(dir, version string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// The store is never touched while generating docs.
	root := cli.NewRoot(version, store.New(os.DevNull), io.Discard, io.Discard)
	root.DisableAutoGenTag = true
	date := time.Now()
	if s := os.Getenv("SOURCE_DATE_EPOCH"); s != "" { // reproducible builds
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			date = time.Unix(n, 0).UTC()
		}
	}
	header := func(section string) *doc.GenManHeader {
		return &doc.GenManHeader{Title: "ALI", Section: section, Source: "ali " + version, Manual: "ali manual", Date: &date}
	}
	if err := doc.GenManTree(root, header("1"), dir); err != nil {
		return err
	}
	// GenManTree skips help topics (they are not runnable); give each a section 7 page.
	for _, c := range root.Commands() {
		if !c.IsAdditionalHelpTopicCommand() {
			continue
		}
		f, err := os.Create(filepath.Join(dir, "ali-"+c.Name()+".7"))
		if err != nil {
			return err
		}
		err = doc.GenMan(c, header("7"), f)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
	}
	return nil
}
