# 0025. Escape control characters when printing stored text and reject NUL

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Security audit 2026-11-08

## Context

`ali ls`, `search`, the `add`/`update` confirmations and the import report printed stored commands raw. A command or a hand-edited `aliases.toml` containing ESC, OSC or C1 control characters could clear the screen, retitle the terminal, or hide text (for example `ESC[8m`), so what a user sees in `ali ls` could differ from what the alias runs. A NUL byte cannot be represented in a zsh alias and silently truncates it.

## Decision

We will print all untrusted text through `display()`, which shows `\n`, `\t`, `\r` and writes any other C0, DEL or C1 control character as `\xNN`. Stored data is not altered, and `--json` output is already escaped by the encoder. `ValidateCommand` rejects NUL bytes.

## Consequences

What `ali` shows is inert and faithful to what is stored. Commands that legitimately contain control characters (rare) appear with visible escapes in listings.

## Alternatives considered

- **Strip control characters**: hides what is actually stored.
- **Sanitise on write**: changes the user's command; they asked for it verbatim (ADR 0013).
