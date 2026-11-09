# 0046. Enforce Conventional Commits locally and in CI

- **Status**: Accepted
- **Date**: 2026-11-09
- **Spec**: Owner request to enforce the commit rule of the constitution for outside contributors

## Context

The constitution requires Conventional Commits, but nothing checked them: the rule held only because the owner followed it. Pull requests are merged with merge commits (ADR 0045), so every commit of a contributor's pull request lands on `main` unchanged.

## Decision

We will check every commit message with one POSIX shell script, `.githooks/commit-msg`, used in two places: as a local `commit-msg` hook (enabled by `make hooks`, together with the pre-push secret scan) and by a required CI job, `commits (Conventional Commits)`, that runs it on every non-merge commit of a pull request. A subject must be `type(scope)!: subject` with type one of feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert; scope and `!` are optional. Messages git writes itself (`Merge ...`) and `fixup!` / `squash!` commits are accepted.

## Consequences

Contributors learn about the rule from the hook or the failing check, with an example of the format, instead of in review. A pull request with a bad message must be fixed by rewording the commits (`git rebase -i`, then a force-push to the contributor's branch). Dependabot already writes `build(deps): ...`. The script has no dependencies, so it adds nothing to the toolchain.

## Alternatives considered

- **commitlint (Node.js)**: a Node toolchain for one regular expression.
- **Check only the pull request title and squash-merge**: conflicts with merge commits (ADR 0045's merge method) and loses the small commits the constitution asks for.
- **A third-party GitHub Action**: another dependency to pin and trust for a ten-line check.
