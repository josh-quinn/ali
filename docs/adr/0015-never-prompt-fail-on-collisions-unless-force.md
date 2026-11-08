# 0015. Never prompt; fail on collisions unless --force

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-008, Constitution VI, clarifications

## Context

The tool must be scriptable and must not clobber existing aliases by accident.

## Decision

We will keep all commands non-interactive. `ali add` fails if the name exists and points to `ali update`; `--force` overwrites. `ali rm` does not ask for confirmation and prints what it removed.

## Consequences

Scriptable and predictable. Users who keep the store in git can recover removals.

## Alternatives considered

- **Interactive confirmation**: breaks scripts and flow.
