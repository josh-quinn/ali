# 0035. One expansion engine in Go behind thin, gated zsh widgets

- **Status**: Proposed
- **Date**: 2026-11-08
- **Spec**: specs/003-command-expansion research R1-R4

## Context

Expansion rules (tokenising, scoping, precedence, quoting) are subtle and security-relevant. Implementing them in zsh would duplicate logic and be hard to test; calling a process on every keypress would violate Constitution I.

## Decision

We will implement all rules in the Go package `internal/expand` and expose them through `ali expand`. Generated zsh glue keeps tables of trigger words and sub-alias names and calls the binary only when a line can involve an ali expansion; it chains to previously bound widgets so other plugins keep working. Line-editor behaviour is tested by driving a real zsh through `zsh/zpty`.

## Consequences

Single source of truth with full unit tests; zero processes for ordinary typing; ~5-8 ms per expanding keypress. zsh glue still needs careful e2e tests and a manual pass with common plugins.

## Alternatives considered

- **Pure zsh engine**: duplicated, hard-to-test logic.
- **Daemon**: state and lifecycle complexity.
