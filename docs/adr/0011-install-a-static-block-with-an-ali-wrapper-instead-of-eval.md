# 0011. Install a static block with an ali() wrapper instead of eval

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: research R1; Constitution I

## Context

A child process cannot change its parent shell, so changes can only appear in the current terminal if the shell itself re-sources the file. The first plan used `eval "$(ali init zsh)"`, which launches `ali` on every new terminal.

## Decision

We will have `ali init zsh` append a block, delimited by `# >>> ali >>>` and `# <<< ali <<<`, that sources `aliases.zsh` and defines a function `ali()` which runs `command ali "$@"` and re-sources the file. Nothing runs the binary at shell startup. Measured startup cost: about 3.8 ms at 1,000 aliases.

## Consequences

Changes are live immediately. The wrapper is zsh syntax, so supporting other shells needs their own block. Because functions are parsed when defined, aliasing builtins the wrapper uses could hijack it (addressed in ADR 0021).

## Alternatives considered

- **`eval "$(ali init zsh)"`**: extra process per shell (~10-30 ms).
- **precmd hook polling mtime**: more machinery.
- **Ask users to `exec zsh`**: breaks flow.
