# 0019. Limit v1 scope to aliases on zsh (macOS first)

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: spec Assumptions

## Context

The owner uses macOS and zsh and wants a focused tool; every extra shell or object type multiplies the work and the test matrix.

## Decision

We will support aliases only (no functions, env vars or per-directory aliases) and zsh only; `ali init` accepts only `zsh`. macOS is the tested platform (CI runs on macOS). The code builds on Linux and other Unixes but is untested there, and does not build on Windows because of `syscall.Flock`.

## Consequences

Clear, small surface. Bash, fish, Linux and Windows users are not served yet; see the README's Platform support section.

## Alternatives considered

- **Support bash and fish now**: needs a per-shell renderer and block, and a larger test matrix.
