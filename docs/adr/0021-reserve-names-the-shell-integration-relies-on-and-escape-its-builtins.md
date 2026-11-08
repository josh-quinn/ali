# 0021. Reserve names the shell integration relies on and escape its builtins

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Security audit 2026-11-08

## Context

Functions are parsed when defined, and aliases expand during parsing. If a user (or ali) defines an alias named `command`, `source`, `local`, `return` or `unset`, the `ali()` wrapper in `.zshrc` is rewritten at definition time and runs the alias instead of the builtin. Reproduced: with `alias command='echo HIJACKED'` every `ali` call printed `HIJACKED ali ...`. Aliasing `source` similarly redirected the generated-file load. Aliasing `ali` itself shadows the wrapper.

## Decision

We will (1) reject the names `ali`, `alias`, `unalias`, `builtin`, `command`, `source`, `local`, `return`, `unset`, `for`, `do`, `done` in `alias.ValidateName` (so `add`, `import` and hand-edited stores all enforce it), and (2) write every builtin in generated code and in the wrapper with a leading backslash (`\source`, `\command ali`, `\alias`, `\unalias`), which stops alias expansion, and (3) single-quote names inside `_ali_names`. Layer 2 also protects against aliases the user defined elsewhere in their rc files, which layer 1 cannot see.

## Consequences

A hijack needs neither a hostile store nor privileges, only an unlucky alias name, so this is a robustness and integrity fix more than a remote vulnerability. Users lose the ability to alias those twelve words. Reserved words `for`, `do` and `done` cannot be backslash-escaped, hence the name ban.

## Alternatives considered

- **Only the name ban**: misses aliases defined outside ali.
- **Only escaping**: aliasing `ali` itself still shadows the wrapper.
