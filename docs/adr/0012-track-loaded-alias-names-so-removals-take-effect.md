# 0012. Track loaded alias names so removals take effect

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: research R2

## Context

Sourcing a file can add or change aliases but cannot remove one that no longer appears.

## Decision

We will have `aliases.zsh` begin by unaliasing every name listed in the array `_ali_names` (set by the previous load) and then redefine the current set and reset `_ali_names`.

## Consequences

`ali rm` and renames work in the live shell. The user's own aliases are never unaliased.

## Alternatives considered

- **`unalias -a`**: destroys the user's aliases.
- **Print an `unalias` hint**: breaks the "live immediately" goal.
