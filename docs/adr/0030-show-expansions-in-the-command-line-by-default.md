# 0030. Show expansions in the command line by default

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: specs/003-command-expansion; owner decision 1

## Context

Sub-aliases, placeholders and environments turn a short line into a different command. Hidden expansion makes it hard to see what will run, pollutes history with unreadable shorthand, and is risky against production.

## Decision

We will expand in the zsh line editor by default: sub-aliases on Space, the whole line on Enter, so the full command is visible before or as it runs and is what history records. A machine-local setting `expand = runtime` switches to hidden expansion at execution for users who prefer short lines.

## Consequences

Transparent and safe by default; copy-pasting from history works. Requires ZLE widgets (complexity tracked in the plan). Run-time mode cannot expand lines that start with a real tool name.

## Alternatives considered

- **Run-time only**: opaque; rejected by the owner as a default.
- **Visible only**: the owner wanted the option.
