# 0014. Write atomically under an advisory lock

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-015, research R5

## Context

Two `ali` processes, a crash or a full disk must not lose or corrupt data.

## Decision

We will do read-modify-write while holding `flock` on `aliases.toml.lock`, writing a temp file in the same directory, fsyncing, then renaming. If the callback fails or the store is invalid, nothing on disk changes.

## Consequences

No lost updates (tested with 12 concurrent writers). Unix-only because of `flock`.

## Alternatives considered

- **No locking**: lost updates.
- **Lock via O_EXCL file**: stale locks after a crash.
