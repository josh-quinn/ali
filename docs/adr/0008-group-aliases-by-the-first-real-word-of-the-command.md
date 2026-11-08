# 0008. Group aliases by the first real word of the command

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-002, research R6

## Context

Users want aliases organised by tool without typing a group name each time.

## Decision

We will infer the group as the lower-cased base name of the first word of the command, skipping `sudo`, `env`, `command`, `time`, `nohup`, `exec`, their flags, and `VAR=value` assignments; an empty result is `misc`. An explicit group (`-g`, `regroup`) is "pinned" and is not re-inferred by `update`.

## Consequences

Predictable and zero-typing. Tools that are really one family (`kubectl`, `k9s`) land in different groups unless regrouped. No built-in tool map to maintain.

## Alternatives considered

- **Known-tool map**: smarter, but more to maintain (rejected as YAGNI).
- **Always explicit**: loses the automatic organisation the user asked for.
