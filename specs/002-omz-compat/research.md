# Research: Oh My Zsh compatibility

## R1: How the binary learns what a name means in the live shell
- **Decision**: For `ali add <name>`, the wrapper sets `ALI_SHELL_DEF` for that single invocation to `alias<TAB><text>`, `function`, or empty, computed with zsh's `$aliases` and `$functions` parameters and excluding names in `_ali_names`. For `ali doctor`, the wrapper writes NUL-separated name/value pairs to the binary's stdin with `--shell-state -`. The data comes from a one-shot `precmd` hook in the block that runs at the first prompt, after every rc file, and takes a bulk copy of the alias table (`_ali_at_start`) plus the generation ID of the `aliases.zsh` it loaded (`_ali_gen`). `doctor` compares the snapshot with the store in Go, and skips the check if the generation no longer matches the store. Capturing at the first prompt is necessary: every later `ali` command re-sources ali's aliases, which would mask redefinitions if state were read when `doctor` runs (found while implementing T012/T013).
- **Rationale**: Only the shell knows its aliases. An env var scoped to one command does not leak; stdin avoids size limits for 1,000 names.
- **Alternatives**: run `zsh -i -c alias` from the binary (slow, re-runs the user's rc files, side effects); parse `.zshrc` and Oh My Zsh plugin files (fragile, misses dynamic definitions).

## R1b: Startup cost (measured while implementing T018)
- A first version looped over every alias at startup (record originals, compare with `_ali_expected`). With 1,000 aliases it added ~7 ms to interactive startup, against the 1 ms budget (SC-005).
- **Decision**: use zsh array set operations instead: restore only `old minus new` names; record originals only for `(existing aliases ∩ new) minus old`; replace the comparison loop with one bulk copy of `$aliases` at the first prompt and do the comparison in Go. Profiled: ~0.4 ms for the intersection and ~0.45 ms for the copy, versus ~2 ms and ~1.9 ms for the loops.
- **Result** (interactive startup, median of 4 x 40 runs, `main` vs this branch): 100 aliases +0.59 ms; 1,000 aliases +1.09 ms. SC-005 is met at typical sizes and missed by ~0.1 ms at 1,000 aliases; the remaining cost is the snapshot copy and one intersection, both needed for correctness.

## R2: Restoring overridden aliases
- **Decision**: `aliases.zsh` keeps `typeset -gA _ali_prev`. On each load it first restores, for every previously loaded ali name, `_ali_prev[name]` if set or unaliases it otherwise; then, for each name about to be defined, records the current alias into `_ali_prev` if one exists and none is recorded; then defines ali's aliases. Assignment uses the `aliases` parameter, not the `alias` builtin.
- **Rationale**: Pure zsh, session-scoped, no stored foreign state (spec clarification). Functions need no handling: removing an alias re-exposes a function automatically.
- **Open check for implementation**: confirm the `zsh/parameter` module (`$aliases`, `$functions`) is available in `zsh -i -d` sessions; otherwise `zmodload -F zsh/parameter` first.

## R3: Detecting block order relative to Oh My Zsh
- **Decision**: `doctor` scans `.zshrc` for a non-comment line containing `oh-my-zsh.sh` and compares its line number with the block start marker.
- **Rationale**: Matches the standard installer's `source $ZSH/oh-my-zsh.sh`. Other frameworks are covered by the live redefinition check (R1).

## R4: Upgrading the block
- **Decision**: The block gains a version line (`# ali block v2, managed by ali init`). `Install` finds start and end markers; if both exist once and the content differs from the current block, it backs up the file and replaces only that span. Zero markers means append; unbalanced or duplicated markers is an error.
- **Rationale**: Byte-exact preservation outside the block (SC-004) and safe failure on damage.
