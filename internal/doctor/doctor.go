// Package doctor checks that ali is wired into the user's shell correctly.
// It is pure: the CLI gathers the inputs and prints the results.
package doctor

import (
	"bufio"
	"fmt"
	"sort"
	"strings"

	"github.com/josh-quinn/ali/internal/shellconf"
)

// Status of one check.
type Status int

const (
	OK Status = iota
	Fail
	Skipped
)

func (s Status) String() string { return [...]string{"ok", "FAIL", "skipped"}[s] }

// Check names, in the order they are reported.
const (
	CheckBlockPresent = "block present in .zshrc"
	CheckBlockCurrent = "block up to date"
	CheckOrder        = "block loads after Oh My Zsh"
	CheckStore        = "aliases.toml readable"
	CheckRedefined    = "aliases not redefined after the block"
)

// Check is one result line.
type Check struct {
	Name   string
	Status Status
	Detail string
	Fix    string
}

// Input is everything doctor looks at.
type Input struct {
	RC        string // contents of .zshrc
	RCMissing bool
	StoreErr  error
	// Overridden maps ali alias names to what the shell had at its first
	// prompt when that differed from ali's definition ("" = not an alias).
	// nil means doctor did not run through the shell wrapper.
	Overridden map[string]string
	// SnapshotStale means the aliases changed since this terminal started, so
	// its startup snapshot cannot be compared with the store.
	SnapshotStale bool
}

// Run performs every check.
func Run(in Input) []Check {
	state, blockLine, _ := shellconf.Inspect(in.RC)
	var cs []Check

	present := Check{Name: CheckBlockPresent, Status: OK}
	switch {
	case in.RCMissing || state == shellconf.BlockMissing:
		present = Check{Name: CheckBlockPresent, Status: Fail, Detail: "ali's block is not in your .zshrc", Fix: "run: ali init zsh"}
	case state == shellconf.BlockDamaged:
		present = Check{Name: CheckBlockPresent, Status: Fail, Detail: "the block's start/end markers are missing, duplicated or out of order", Fix: "fix the '# >>> ali >>>' / '# <<< ali <<<' lines by hand, then run: ali init zsh"}
	}
	cs = append(cs, present)

	current := Check{Name: CheckBlockCurrent, Status: OK}
	switch {
	case present.Status != OK:
		current = Check{Name: CheckBlockCurrent, Status: Skipped, Detail: "no usable block"}
	case state == shellconf.BlockOutdated:
		current = Check{Name: CheckBlockCurrent, Status: Fail, Detail: "the block was written by an older ali", Fix: "run: ali init zsh (replaces only the block, with a backup)"}
	}
	cs = append(cs, current)

	order := Check{Name: CheckOrder, Status: OK}
	omzLine := ohMyZshLine(in.RC)
	switch {
	case present.Status != OK:
		order = Check{Name: CheckOrder, Status: Skipped, Detail: "no usable block"}
	case omzLine == 0:
		order = Check{Name: CheckOrder, Status: Skipped, Detail: "Oh My Zsh is not sourced in .zshrc"}
	case omzLine > blockLine:
		order = Check{Name: CheckOrder, Status: Fail,
			Detail: fmt.Sprintf("oh-my-zsh.sh is sourced on line %d, after ali's block on line %d, so its aliases replace ali's", omzLine, blockLine),
			Fix:    "move the ali block below the 'source $ZSH/oh-my-zsh.sh' line"}
	}
	cs = append(cs, order)

	st := Check{Name: CheckStore, Status: OK}
	if in.StoreErr != nil {
		st = Check{Name: CheckStore, Status: Fail, Detail: in.StoreErr.Error(), Fix: "fix the file at the position shown, or restore it from your dotfiles"}
	}
	cs = append(cs, st)

	red := Check{Name: CheckRedefined, Status: OK}
	switch {
	case in.Overridden == nil:
		red = Check{Name: CheckRedefined, Status: Skipped, Detail: "run 'ali doctor' from your shell (not a script) so ali can see what your shell loaded"}
	case in.SnapshotStale:
		red = Check{Name: CheckRedefined, Status: Skipped, Detail: "your aliases changed since this terminal started; open a new terminal and run 'ali doctor' there"}
	case len(in.Overridden) > 0:
		names := make([]string, 0, len(in.Overridden))
		for n := range in.Overridden {
			names = append(names, n)
		}
		sort.Strings(names)
		parts := make([]string, len(names))
		for i, n := range names {
			now := in.Overridden[n]
			if now == "" {
				now = "not an alias"
			}
			parts[i] = fmt.Sprintf("%s (now: %s)", n, now)
		}
		red = Check{Name: CheckRedefined, Status: Fail,
			Detail: "something loaded after ali's block changed: " + strings.Join(parts, ", "),
			Fix:    "remove those definitions from later in .zshrc, or keep ali's block last"}
	}
	return append(cs, red)
}

// Failed reports whether any check failed.
func Failed(cs []Check) bool {
	for _, c := range cs {
		if c.Status == Fail {
			return true
		}
	}
	return false
}

// ohMyZshLine returns the 1-based line that sources oh-my-zsh.sh, or 0.
func ohMyZshLine(rc string) int {
	sc := bufio.NewScanner(strings.NewReader(rc))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "#") && strings.Contains(line, "oh-my-zsh.sh") {
			return n
		}
	}
	return 0
}

// ParseState decodes the wrapper's startup snapshot: the generation of the
// aliases.zsh the shell loaded, then NUL-separated name/value pairs of every
// alias the shell had at its first prompt. The map is never nil, because
// receiving state at all means the wrapper ran.
func ParseState(s string) (gen string, snapshot map[string]string) {
	snapshot = map[string]string{}
	parts := strings.Split(s, "\x00")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return "", snapshot
	}
	gen, parts = parts[0], parts[1:]
	for i := 0; i+1 < len(parts); i += 2 {
		if parts[i] != "" {
			snapshot[parts[i]] = parts[i+1]
		}
	}
	return gen, snapshot
}

// Compare returns, for each ali alias whose snapshot value differs from the
// store, what the shell had instead ("" = not an alias at all).
func Compare(store, snapshot map[string]string) map[string]string {
	diff := map[string]string{}
	for name, cmd := range store {
		if live, ok := snapshot[name]; !ok || live != cmd {
			diff[name] = live
		}
	}
	return diff
}
