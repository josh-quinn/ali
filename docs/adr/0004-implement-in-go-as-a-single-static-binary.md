# 0004. Implement in Go as a single static binary

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: plan.md Technical Context

## Context

`ali` runs many times a day and must start instantly (Constitution I). It will be distributed publicly, ideally via Homebrew and GitHub Releases, to users who may not have a language runtime.

## Decision

We will write `ali` in Go (module `github.com/josh-quinn/ali`, Go 1.24+) and ship a statically linked binary (`CGO_ENABLED=0`).

## Consequences

Measured: about 5-8 ms per command at 1,000 aliases. Cross-compilation is trivial. Contributors need Go. The store code uses `syscall.Flock`, so the code is Unix-only (see ADR 0019).

## Alternatives considered

- **Python**: slower startup, runtime dependency for users.
- **Rust**: same benefits but cargo was absent and the gain is small here.
- **Shell script**: hard to test and maintain publicly.
