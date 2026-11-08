# 0020. Test end to end through a line-fed interactive zsh

- **Status**: Accepted
- **Date**: 2026-11-08

## Context

`zsh -c "<script>"` parses the whole script before running it, so an alias defined on one line does not expand on the next. An early version of the end-to-end tests passed falsely for that reason (the assertion matched `ali add`'s own confirmation text).

## Decision

We will drive end-to-end tests by piping commands into `zsh -i -d` one line at a time in a throwaway `HOME`/`ZDOTDIR` (with a quiet prompt set in a sandbox `.zshenv`), and assert on exact output lines.

## Consequences

Tests reflect what a person typing experiences, including live update and removal. They require `zsh` and are skipped without it.

## Alternatives considered

- **`zsh -c`**: misleading results as described.
- **Mock the shell**: would not catch parse-time alias behaviour.
