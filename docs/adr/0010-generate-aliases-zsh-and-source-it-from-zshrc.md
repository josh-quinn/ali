# 0010. Generate aliases.zsh and source it from .zshrc

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: research R1

## Context

Normal commands must not edit the user's `.zshrc`, yet new shells must see the aliases.

## Decision

We will render `aliases.zsh` from the store on every change. `.zshrc` gets one managed block (ADR 0011) that sources it. Only `ali init` and `ali import` touch user files.

## Consequences

The generated file is derived and reproducible; editing it by hand is pointless and the header says so.

## Alternatives considered

- **Append to `.zshrc` directly**: the problem being solved.
- **`eval "$(ali ...)"` at startup**: see ADR 0011.
