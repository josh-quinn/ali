# 0033. Use @ for call-time values and @@ for a one-off environment

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: owner decision 4

## Context

Values must be typed inside the user's command line. Flags like `-ga` collide with the tool's own flags, with `ali add -g`, and with combined short flags. Real commands also use `@` (`npm i @types/node`, `curl -d @file`, git `@{u}`).

## Decision

We will use `@key` to pick a named value for the only placeholder whose variable has that key, `@var=value` to set a variable (named value first, literal otherwise), and `@@env` to use an environment for one line. An `@` word is consumed only if it matches a placeholder in that line; all others are left untouched. Ambiguous `@key` stops the line with a message.

## Consequences

Short to type, no clash with tool flags, and existing `@` arguments keep working. A value key that happens to equal a literal argument is only consumed when a placeholder for that variable is present.

## Alternatives considered

- **`-ga`/`-gal` flags**: collide as described.
- **`%` or `:`**: `%` appears in date/printf formats, `:` in URLs and kubectl resources.
