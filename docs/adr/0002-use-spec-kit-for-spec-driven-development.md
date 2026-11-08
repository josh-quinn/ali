# 0002. Use Spec Kit for spec-driven development

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: specs/001-alias-manager/

## Context

The project is meant to go public and should show how requirements became code. The owner asked for the full Spec Kit flow.

## Decision

We will develop with Spec Kit (`specify init --integration claude`): constitution, specify, clarify, plan, tasks, analyze, implement. Artifacts live in `.specify/` and `specs/NNN-name/`. `.specify/feature.json` is per-checkout state and is not committed, as Spec Kit's own `.gitignore` directs.

## Consequences

Requirements, design and tasks are versioned beside the code. Contributors must keep specs and code in step. The Spec Kit slash commands were not visible to the session that bootstrapped the project, so their instructions were followed by hand.

## Alternatives considered

- **Ad-hoc design in issues**: no durable trace.
- **Superpowers brainstorming only**: used for the initial design dialogue, but produces no repo artifacts on its own.
