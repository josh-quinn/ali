# 0038. Generate help and man pages from the command tree

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Constitution v1.2.0; owner review of features 002/003

## Context

The owner asked for the tool to be well documented with help and a man page. The CLI is defined with cobra, whose `cobra/doc` package can render man pages from the same command definitions that produce `--help`, so the two cannot drift. Feature 003 adds syntax typed in the shell (`\word`, `{{name}}`, `@`, `@@`) that is not a subcommand and so has no `--help` of its own.

## Decision

We will give every command a Short, Long and Example; add help topics (`ali help expansion`, `ali help placeholders`, `ali help environments`, `ali help files`) as cobra help-topic commands; and generate `man/ali.1`, one page per subcommand and pages for the topics with a small build-only program (`internal/tools/gendocs`, run by `make man` and by GoReleaser before packaging). The Homebrew formula installs the pages. A unit test walks the command tree and fails if any command lacks Short, Long or Example. `go install` users can run `ali help` or generate pages with `make man`.

## Consequences

Help and man pages always match the binary. The gendocs program pulls in `github.com/cpuguy83/go-md2man/v2` (already in the module graph as cobra's own documentation dependency); it is used only by the build-time program and is not linked into `ali` (Constitution V). Writing good Long/Example text becomes part of every command task.

## Alternatives considered

- **Hand-written man page**: drifts from `--help`.
- **Help only, no man page**: the owner asked for a man page.
- **Ship man pages via `ali man` at run time**: adds docs code to the binary for little gain.
