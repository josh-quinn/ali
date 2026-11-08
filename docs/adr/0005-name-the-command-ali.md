# 0005. Name the command ali

- **Status**: Accepted
- **Date**: 2026-11-08

## Context

The command is typed constantly, so it should be short, and it needs to be distinct enough to publish.

## Decision

We will call the executable `ali`.

## Consequences

Two-letter-ish and memorable. The `.zshrc` wrapper function is also named `ali`, which is why the name `ali` is reserved from being aliased (see ADR 0022). Should be checked for collisions in Homebrew before the first release.

## Alternatives considered

- **al**: shorter but more likely to collide.
- **alias-vault**: descriptive but long to type.
