# Research: Command expansion

## R1: Where the expansion logic lives
- **Decision**: All rules live in Go (`internal/expand`). zsh code is glue: it decides cheaply *whether* to ask, then calls `ali expand` and applies the result.
- **Rationale**: One implementation, fully unit-testable (Constitution IV); zsh string handling is error-prone for quoting rules.
- **Alternatives**: implement expansion in zsh from generated associative arrays (no process at all, but duplicated logic and hard to test); a long-running daemon (complexity, state).

## R2: Avoiding a process for ordinary typing
- **Decision**: `aliases.zsh` defines `_ali_trigger`, an associative array of first words that can start an expandable line: every tool with sub-aliases, every ali alias whose tool has sub-aliases, and every ali alias containing a placeholder (plus the wrappers `sudo`, `env`, `time`, `nohup`, `command`, which are looked through). The Space widget calls the binary only if the line's tool word is in the table and the word just finished is in the tool's sub list (also generated) or starts with `@`; Enter calls it only for trigger lines.
- **Rationale**: Meets Constitution I: zero processes for everything else.
- **Measured baseline**: one `ali` run is ~5-8 ms on the dev machine; budget is 15 ms.

## R3: Coexisting with other line-editor plugins
- **Decision**: Bind Space and Enter in the `main` and `viins` keymaps to `_ali_space` / `_ali_accept`, which do their work and then call the widget that was bound before (captured at load: `zle -A accept-line _ali_orig_accept`, likewise for `magic-space`/`self-insert`). Load order: the ali block is last in `.zshrc`, so it wraps whatever Oh My Zsh and others installed. Expansion is skipped while `$PENDING`/bracketed paste is active.
- **Rationale**: The approach used by zsh-abbr; preserves autosuggestions and syntax highlighting.
- **Risk**: zsh-syntax-highlighting re-wraps widgets when it loads; if loaded after ali it still calls ali's widgets, which is fine. Verify manually with zsh-autosuggestions and zsh-syntax-highlighting before release (no CI coverage).

## R4: Protocol between the widget and the binary
- **Decision**: `ali expand --zle --cursor N -- "$BUFFER"` prints NUL-separated fields: status, new cursor, new buffer, message. Status `ok`, `unchanged`, `confirm` (protected environment: show the resolved line, require a second Enter), `error` (message shown with `zle -M`, line not run).
- **Rationale**: NUL cannot appear in a command line, so no escaping is needed.

## R5: Run-time mode
- **Decision**: With `expand = runtime`, aliases that can expand are rendered as functions: `k() { local _c; _c="$(\command ali expand --run k "$@")" || \return $?; \eval "$_c"; }`. `ali expand --run` re-quotes every word it outputs, so `eval` only ever sees ali-quoted words. Lines that start with a tool name (`kubectl g p`) are not expanded in this mode (spec US4 #2), because defining functions named after real tools would hide them. Protected environments prompt on `/dev/tty`.
- **Alternatives**: global aliases (`alias -g`) for sub-aliases (expand everywhere, dangerous); a `preexec` hook (cannot change what runs).

## R6: Tokenising the command line
- **Decision**: A shell-word tokenizer that records, per word, its raw text span, whether any part was quoted or escaped, and whether it is a plain word. Only plain unquoted words can be aliases, sub-aliases or `@` words; everything else is copied byte-for-byte. Words after an unquoted `;`, `|`, `&&`, `||` or `&` start a new command and are expanded independently.
- **Rationale**: Preserves the user's text exactly where ali does not act.

## R7: Safe insertion of values
- **Decision**: A substituted value is emitted as-is if it matches `^[A-Za-z0-9_./:=,%+@-]+$`, otherwise single-quoted with `'\''` escaping (ADR 0013). Placeholders must be outside quotes in definitions (validated at save time by the same tokenizer), so a value is always inserted in unquoted context. A placeholder inside a larger word (`--namespace={{ns}}`) concatenates correctly because quoting is per value.
- **Rationale**: Constitution VII; values are data.

## R8: Per-terminal and global environments
- **Decision**: The `ali()` wrapper handles `ali env use <name>` and `ali env off`: it runs the binary to validate, then exports or unsets `ALI_ENV` in the current shell. `--global` makes the binary write `global-env` (machine-local) and the wrapper unsets `ALI_ENV`, so the current terminal follows the global choice too. Resolution in the binary: `@@env` in the line, then `ALI_ENV`, then `global-env`.
- **Rationale**: A child process cannot set its parent's variables (same reason as ADR 0011); exported so `ali expand` sees it.

## R9: Prompt indicator without processes
- **Decision**: A `precmd` hook (via `add-zsh-hook`) computes `_ali_env_label` from `ALI_ENV` or the contents of `global-env` read with zsh's `$(<file)` (no fork), and sets `RPROMPT` to the label followed by the user's original `RPROMPT` (captured once). Protected environments are listed in the generated file and shown with `%F{red}`. `prompt = off` disables it; `ali_prompt_info` is provided for custom themes. ali never runs `setopt prompt_subst`.
- **Rationale**: Visible (owner decision 3) at zero process cost; works with robbyrussell, which leaves `RPROMPT` empty.

## R10: Testing the line editor
- **Decision**: Drive a real interactive zsh inside `zsh/zpty` from a test script; send keystrokes (`k g p `, Enter), read the rendered line and the command output. The Go e2e test runs that zsh script in the sandbox `HOME`.
- **Rationale**: ZLE is inactive when stdin is not a terminal, so the line-fed approach of ADR 0020 cannot test widgets. `zpty` ships with zsh, so no new dependency.
- **Verified 2026-11-08** (throwaway probe, not committed): on macOS system zsh 5.9, a `zpty`-driven interactive zsh with a Space widget turned `echo g x` into `echo get x` on screen and printed `get x`. Output includes bracketed-paste escape sequences, which the test matcher must tolerate.
- **Still open**: confirm on GitHub's macOS runners (first CI run).

## R11: Tab completion (P3)
- **Decision**: Generate a completion function into `~/.config/ali/completions/_ali` (ali's own commands) and a Tab widget that, when the current word starts with `@` in a trigger line, offers value names via `compadd`, otherwise falls through to the previously bound completion widget. Sub-aliases after a tool are offered the same way when the word is in sub-alias position.
- **Risk**: highest-integration-risk part; deliberately last (P3).

## R12: Store compatibility
- **Decision**: Write `version = 1` when no feature-003 tables are present, `version = 2` otherwise. Older ali versions reject v2 with "unsupported version" (existing behaviour), so a synced store is never corrupted by an old binary.

## R13: Measured performance (T031, 2026-11-08, dev machine, 1,000 aliases + sub-aliases, a variable and an environment)
- **Per expanding keypress** (one `ali expand` run): Space 5.35 ms median / 5.81 ms p95; Enter 5.37 / 5.81; Tab 5.27 / 5.79. Budget 15 ms (SC-003): met. Lines with no trigger word start no process (e2e-tested).
- **Startup, first version**: the widget setup captured the previous Space and Tab bindings with `$(bindkey ...)` (a fork of the whole shell, ~3.5-4 ms each) and loaded `zsh/complete` (~2 ms). Piped `zsh -i` startup was +13.3 ms for all of ali.
- **Decision**: wrap the widgets (`self-insert`, `magic-space`, the completion widgets, `accept-line`) instead of rebinding keys, as zsh-autosuggestions does, so no binding needs to be read; load `zsh/complete` lazily on the first Tab that has ali suggestions. A test forbids command substitution and `zsh/complete` at source time.
- **Result**: piped `zsh -i` (no line editor): +9.4 ms for all of ali, of which ~4 ms is loading `zsh/zle` triggered by wrapping `accept-line`, which a real terminal loads anyway. In a real pseudo-terminal (zpty, line editor active): ~+1.3 ms for all of ali versus no ali. The pty timing loop is coarser than the piped one, so treat it as approximate.
