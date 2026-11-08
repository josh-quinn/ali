// Package shellconf edits the user's .zshrc: only the managed ali block, only
// on request, always after a timestamped backup.
package shellconf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/josh-quinn/ali/internal/render"
	"github.com/josh-quinn/ali/internal/store"
)

// Path returns the .zshrc location, honouring an absolute ZDOTDIR. It fails
// rather than return a relative path, which would edit a .zshrc in whatever
// directory the process is in.
func Path() (string, error) {
	if z := os.Getenv("ZDOTDIR"); filepath.IsAbs(z) {
		return filepath.Join(z, ".zshrc"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil || !filepath.IsAbs(home) {
		return "", fmt.Errorf("cannot locate .zshrc: HOME is not set")
	}
	return filepath.Join(home, ".zshrc"), nil
}

// Result describes what Install did.
type Result struct {
	Changed bool
	Updated bool   // an existing, outdated block was replaced
	Backup  string // empty when no backup was needed
}

// BlockState says whether, and in what shape, an rc file contains the ali block.
type BlockState int

// Block states returned by Inspect.
const (
	BlockMissing  BlockState = iota // no markers
	BlockCurrent                    // exactly one block, identical to render.Block()
	BlockOutdated                   // exactly one well-formed block with other content
	BlockDamaged                    // markers missing, duplicated or out of order
)

func (s BlockState) String() string {
	return [...]string{"missing", "current", "outdated", "damaged"}[s]
}

// HasBlock reports whether content already contains the ali block.
func HasBlock(content string) bool { return strings.Contains(content, render.BlockStart) }

// Inspect locates the ali block. For a well-formed block it also returns its
// 1-based start line and its byte span [from, to) including the end marker's
// newline, so callers can replace exactly that region.
func Inspect(content string) (state BlockState, startLine int, span [2]int) {
	starts := strings.Count(content, render.BlockStart)
	ends := strings.Count(content, render.BlockEnd)
	if starts == 0 && ends == 0 {
		return BlockMissing, 0, span
	}
	if starts != 1 || ends != 1 {
		return BlockDamaged, 0, span
	}
	from := strings.Index(content, render.BlockStart)
	endAt := strings.Index(content, render.BlockEnd)
	if endAt < from {
		return BlockDamaged, 0, span
	}
	to := endAt + len(render.BlockEnd)
	if to < len(content) && content[to] == '\n' {
		to++
	}
	span = [2]int{from, to}
	startLine = strings.Count(content[:from], "\n") + 1
	if content[from:to] == render.Block() {
		return BlockCurrent, startLine, span
	}
	return BlockOutdated, startLine, span
}

// Backup copies path to <path>.ali-backup-<timestamp>, preserving its mode.
func Backup(path string) (string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- reading the user's own config file is the purpose
	if err != nil {
		return "", err
	}
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	dst := path + ".ali-backup-" + time.Now().Format("20060102-150405.000000000")
	return dst, store.WriteAtomic(dst, data, st.Mode().Perm())
}

// Install adds the ali block to the rc file, or replaces an outdated one in
// place (ADR 0029). Content outside the block is never changed; a damaged block
// is refused without writing anything.
func Install(path string) (Result, error) {
	cur, err := os.ReadFile(path) // #nosec G304 -- reading the user's own config file is the purpose
	if err != nil && !os.IsNotExist(err) {
		return Result{}, err
	}
	exists := err == nil
	state, _, span := Inspect(string(cur))
	var body string
	switch state {
	case BlockCurrent:
		return Result{}, nil
	case BlockDamaged:
		return Result{}, fmt.Errorf("%s: the ali block markers (%q ... %q) are missing, duplicated or out of order; fix or remove them by hand, then run ali init zsh again", path, render.BlockStart, render.BlockEnd)
	case BlockOutdated:
		body = string(cur[:span[0]]) + render.Block() + string(cur[span[1]:])
	case BlockMissing:
		body = string(cur)
		if body != "" && !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		if body != "" {
			body += "\n"
		}
		body += render.Block()
	}
	res := Result{Updated: state == BlockOutdated}
	mode := os.FileMode(0o644)
	if exists {
		if res.Backup, err = Backup(path); err != nil {
			return Result{}, err
		}
		if st, serr := os.Stat(path); serr == nil {
			mode = st.Mode().Perm()
		}
	}
	if err := store.WriteAtomic(path, []byte(body), mode); err != nil {
		return Result{}, err
	}
	res.Changed = true
	return res, nil
}
