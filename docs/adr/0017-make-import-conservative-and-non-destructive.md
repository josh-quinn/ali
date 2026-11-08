# 0017. Make import conservative and non-destructive

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: FR-016, research R8, clarifications

## Context

Users already have aliases in `.zshrc`; a wrong import could change meaning or damage the file.

## Decision

We will import only single-line `alias name=value` definitions the parser can read with certainty. Double-quoted values containing `$` or backticks, `alias -g/-s`, multiple definitions on one line and multi-line values are skipped and reported. The source file is never modified; originals are listed so the user can delete them.

## Consequences

Safe, but some real-world aliases need manual conversion.

## Alternatives considered

- **Remove originals automatically**: risky.
- **Execute the rc file and read `alias` output**: runs arbitrary user code.
