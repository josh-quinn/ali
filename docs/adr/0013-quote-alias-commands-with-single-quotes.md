# 0013. Quote alias commands with single quotes

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: research R3, SC-006

## Context

Commands may contain quotes, `$`, backticks, `&&`, pipes, newlines and unicode and must run exactly as typed.

## Decision

We will emit `alias -- name='...'`, replacing each `'` with `'\''`. Names are validated instead of quoted.

## Consequences

Total, reversible transform; verified by round-tripping through real zsh in tests.

## Alternatives considered

- **`$'...'` quoting**: more escapes, higher bug risk.
- **`printf %q`**: output is shell-dependent.
