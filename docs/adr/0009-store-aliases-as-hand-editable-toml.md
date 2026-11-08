# 0009. Store aliases as hand-editable TOML

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-014, data-model.md, research R4

## Context

Users want to read, diff, sync and hand-edit their aliases.

## Decision

We will persist aliases in `$ALI_CONFIG_DIR` or `$XDG_CONFIG_HOME/ali` or `~/.config/ali` as `aliases.toml` with `version = 1`, one table per group (`[helm]` then `hgp = "helm get pods"`), and a `[_pinned] names = [...]` table for explicit groups. Invalid files are reported with their path and never overwritten.

## Consequences

Friendly to dotfile repos. Group names `version` and `_pinned` are reserved. A name may appear in only one group.

## Alternatives considered

- **JSON**: no comments, noisy diffs.
- **YAML**: heavier, ambiguous scalars.
- **A zsh file as the store**: hard to parse safely.
