# 0016. Warn but do not block when an alias shadows a command

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-010, research R7

## Context

Overriding `ls` or `cat` is common and intentional, but silently hiding a command can surprise.

## Decision

We will print a warning to stderr when a new name matches a zsh builtin/keyword or an executable on `PATH`, and still create it. The check is a static list plus a `PATH` scan, not a `zsh` subprocess.

## Consequences

Fast (no process spawn). Aliases, functions and commands defined elsewhere in the user's rc files are not detected.

## Alternatives considered

- **Block creation**: too restrictive.
- **Ask zsh via `whence`**: 30+ ms.
