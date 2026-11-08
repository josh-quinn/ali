# Research: Alias Manager

## R1: Applying changes to the live shell
- **Decision**: `ali init zsh` appends a marked block to `.zshrc`:
  ```zsh
  # >>> ali >>>
  [ -f "${XDG_CONFIG_HOME:-$HOME/.config}/ali/aliases.zsh" ] && source "${XDG_CONFIG_HOME:-$HOME/.config}/ali/aliases.zsh"
  ali() { command ali "$@"; local rc=$?; source "${XDG_CONFIG_HOME:-$HOME/.config}/ali/aliases.zsh" 2>/dev/null; return $rc; }
  # <<< ali <<<
  ```
- **Rationale**: A child process cannot change its parent shell, so a function wrapper that re-sources after each call is required. Writing the static block (instead of `eval "$(ali init zsh)"`) avoids spawning the binary on every shell start (Constitution I).
- **Alternatives**: `eval "$(ali init zsh)"` (extra process per shell, ~10-30 ms); precmd hook polling mtime (more machinery); asking the user to run `exec zsh` (breaks flow).

## R2: Removing aliases from a live shell
- **Decision**: `aliases.zsh` begins by unaliasing every name `ali` previously loaded (tracked in a zsh array `_ali_names`) and then redefines current aliases and resets `_ali_names`.
- **Rationale**: Sourcing alone cannot undefine removed or renamed aliases; tracking only `ali`-owned names never touches the user's other aliases.
- **Alternatives**: `unalias -a` (destroys user aliases); `ali rm` printing an `unalias` hint (breaks SC-002).

## R3: Quoting
- **Decision**: Emit every alias as `alias name='...'`, replacing each `'` with `'\''`. Reject only names, never commands, for special characters.
- **Rationale**: Single-quoted strings in zsh are literal except for `'`, so the transform is total and reversible. Newlines are literal within single quotes.
- **Alternatives**: `$'...'` ANSI-C quoting (more escapes, higher bug risk); `printf %q` (shell-dependent output).

## R4: Storage format
- **Decision**: TOML with one table per group: `[helm]` then `hgp = "helm get pods"`. Explicit-group intent is recorded in a reserved `[_pinned]` table (`names = [...]`) only for aliases whose group was set explicitly (`-g`/`regroup`).
- **Rationale**: Tables give natural grouping and clean diffs; strings need no escaping hacks; `go-toml/v2` is small and fast.
- **Alternatives**: JSON (no comments, noisy diffs); YAML (heavier, ambiguous scalars); a zsh file as the primary store (hard to parse safely).

## R5: Concurrency and atomicity
- **Decision**: Advisory `flock` on `aliases.toml.lock` during read-modify-write; write to a temp file in the same directory, `fsync`, then `rename`.
- **Rationale**: Prevents lost updates between concurrent invocations and torn writes on crash.

## R6: Group inference
- **Decision**: Tokenise the command with a small shell-word splitter; skip leading `sudo`, `env`, `command`, `time`, `nohup`, and `VAR=value` tokens; take the base name of the first remaining token; lowercase it. If nothing remains, group is `misc`.
- **Rationale**: Matches the clarified requirement and handles the common prefixes without a tool map (YAGNI).

## R7: Shadow detection
- **Decision**: Warn if `name` is a zsh builtin/reserved word (static list) or an executable found on `PATH`. Warning goes to stderr; creation proceeds (user explicitly asked for it).
- **Alternatives**: Ask zsh via `zsh -c 'whence'` (slow, 30+ ms); block creation (too restrictive: overriding `ls` is common and intentional).

## R8: Import parsing
- **Decision**: Line-oriented parser for `alias name=value`, `alias -g` is rejected; value may be single-, double-quoted, or bare. Lines it cannot parse safely are reported, not guessed.
- **Rationale**: Never mis-import; unknown constructs (multi-line, conditionals) are surfaced to the user.

## R9: Distribution
- **Decision**: GoReleaser builds darwin arm64/amd64 (plus linux) tarballs with checksums and updates a Homebrew tap formula; CI on `macos-latest` runs vet, gofmt check, and tests.
