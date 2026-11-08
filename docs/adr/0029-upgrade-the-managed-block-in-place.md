# 0029. Upgrade the managed block in place

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: specs/002-omz-compat research R4

## Context

`ali init zsh` currently does nothing when a block exists, so users can never receive changes to the block (needed for ADRs 0027 and 0028).

## Decision

We will version the block (a `# ali block vN` line), and have `ali init zsh` replace the text between the start and end markers when it differs from the current block, after a timestamped backup, leaving every other byte unchanged. Missing end markers or duplicate blocks are errors and nothing is written.

## Consequences

Block changes reach existing users with one command, and `ali doctor` can report an outdated block. ali now edits an existing region of `.zshrc`, so the marker parsing must be strict.

## Alternatives considered

- **Append a second block**: duplicates and order bugs.
- **Ask users to edit `.zshrc` by hand**: error-prone.
