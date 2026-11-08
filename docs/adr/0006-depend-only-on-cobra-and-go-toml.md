# 0006. Depend only on cobra and go-toml

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Constitution V

## Context

Minimal dependencies keep the binary small, the supply chain reviewable and startup fast.

## Decision

We will depend only on `github.com/spf13/cobra` (CLI parsing, help, usage errors) and `github.com/pelletier/go-toml/v2` (store format). Anything else needs justification in a plan's complexity table and an ADR.

## Consequences

Less code to write for flags and TOML; two third-party trees to keep patched (Dependabot is configured).

## Alternatives considered

- **Standard library `flag`**: no subcommand help or ergonomics.
- **urfave/cli, kong**: no advantage over cobra, which is the de-facto standard.
