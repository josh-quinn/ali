# 0018. Back up user files before changing them

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-012, Constitution II

## Context

`.zshrc` is the user's most sensitive dotfile.

## Decision

We will copy it to `<file>.ali-backup-<timestamp>` (same mode) before `ali init` modifies it, and `ali import` backs up the file it reads. Writes are atomic and idempotent, and are limited to the marked block.

## Consequences

Backups accumulate; the user may delete them.

## Alternatives considered

- **No backup**: unacceptable risk.
- **Single rolling backup**: loses earlier states.
