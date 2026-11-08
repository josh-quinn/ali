# 0003. Small, green, conventional commits

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Constitution v1.1.0, Development Workflow

## Context

A public repo is read through its history. Large mixed commits are hard to review, bisect and revert.

## Decision

We will make small single-purpose commits that each pass `make fmt-check vet test`, written as Conventional Commits, with tests committed alongside the code they cover. The original history was rebuilt retrospectively into such commits before anything was pushed.

## Consequences

Any commit can be checked out and tested. Release notes can be generated from messages. Slightly more effort per change.

## Alternatives considered

- **Squash-everything**: loses the story of how the tool was built.
- **Free-form messages**: no tooling support.
