# 0027. Learn about the live shell through the wrapper

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: specs/002-omz-compat research R1

## Context

The binary runs as a child process and cannot see aliases or functions defined in the user's shell (for example the Oh My Zsh git plugin's `gp='git push'`), so `ali add gp ...` overrides it silently. The `ali()` wrapper function does run inside the shell.

## Decision

We will have the wrapper compute what a name means in the live shell and pass it to the binary: `ALI_SHELL_DEF` (alias text, `function`, or empty) for `ali add`, and name/definition lines on stdin for `ali doctor --shell-state -`. Both are internal interfaces between a wrapper and binary of the same version. When the binary runs outside the wrapper it falls back to the builtin/PATH check (ADR 0016).

## Consequences

Oh My Zsh and `.zshrc` aliases become visible to warnings and to `doctor` at no startup cost. The wrapper grows slightly and must stay hijack-safe (ADR 0021). Scripts calling the binary directly get the weaker check.

## Alternatives considered

- **Spawn `zsh -i -c alias` from the binary**: slow and re-runs the user's rc files.
- **Parse `.zshrc` and plugin files**: fragile and misses dynamic definitions.
