# 0036. Insert substituted values as single literal words

- **Status**: Proposed
- **Date**: 2026-11-08
- **Spec**: Constitution VII; Security rule

## Context

Variable values and `@var=value` text come from the store or the command line and may contain spaces, quotes, `;`, `$()` or backticks. Substituting them as raw text would turn data into shell code.

## Decision

We will insert each substituted value as exactly one shell word: unchanged if it consists only of safe characters, otherwise single-quoted (ADR 0013). Placeholders inside quotes in a definition are rejected at save time so insertion always happens in unquoted context. Run-time mode `eval`s only ali-quoted output.

## Consequences

Hostile values cannot execute; tested with a table of payloads (SC-004). A value cannot deliberately expand to several arguments; users write two placeholders or put the words in the definition.

## Alternatives considered

- **Raw substitution**: command injection.
- **Allow quoted placeholders with context-aware escaping**: complex and error-prone.
