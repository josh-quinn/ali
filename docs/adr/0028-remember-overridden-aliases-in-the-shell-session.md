# 0028. Remember overridden aliases in the shell session

- **Status**: Proposed
- **Date**: 2026-11-08
- **Spec**: specs/002-omz-compat research R2

## Context

When an ali alias replaces a pre-existing alias (for example Oh My Zsh's `gp`) and is later removed, the live shell currently has no `gp` at all, while a new shell has the plugin's version again.

## Decision

We will have `aliases.zsh` keep a session-only associative array `_ali_prev` of originals it overrode, restore them when ali's alias disappears, and otherwise unalias as before. Nothing about other tools' aliases is written to the store.

## Consequences

Removal is consistent between the current and new shells. The store stays portable across machines with different plugins. Originals are as of shell load time.

## Alternatives considered

- **Store originals in `aliases.toml`**: leaks machine-specific plugin state into a synced file.
- **Tell the user to open a new shell**: breaks the live-update goal (ADR 0011).
