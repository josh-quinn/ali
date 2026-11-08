# 0032. Use named {{name}} placeholders recognised only for defined variables

- **Status**: Proposed
- **Date**: 2026-11-08
- **Spec**: owner decision 2; spec clarifications

## Context

The original idea addressed arguments by position and list number (`-ga 1=1 2=1`), which is slow to recall and silently wrong if a list is reordered (possibly targeting production). Any placeholder syntax risks colliding with real commands: `<x>` is redirection, `{x}` appears in awk, `find -exec {}` and `${VAR}`, `{{x}}` in Go templates.

## Decision

We will write placeholders as `{{name}}` and treat them as placeholders only when `name` is a defined variable; otherwise the text is literal. Placeholders inside quotes are rejected at save time. Values are referred to by name (`one`, `sys`), never by position.

## Consequences

Readable definitions, order-independent values. A Go template using `{{end}}` breaks only if the user also defines a variable `end`, which `ali var add` can warn about.

## Alternatives considered

- **Positional numbers**: rejected by the owner.
- **`<name>`**: shell redirection.
- **`{name}`**: collides with common commands.
