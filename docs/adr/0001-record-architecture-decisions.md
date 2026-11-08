# 0001. Record architecture decisions

- **Status**: Accepted
- **Date**: 2026-11-08

## Context

Spec Kit captures *what* to build (`spec.md`) and, per feature, *how* (`plan.md`,
`research.md`). It has no project-level, long-lived record of individual architectural
decisions: `research.md` is scoped to one feature and `constitution.md` holds principles,
not choices. Spec Kit core ships no ADR support (a community `adrkit` extension exists but
is discovery-only and unvetted, so we do not depend on it). For a public project, readers
and future contributors need to find out *why* things are the way they are.

## Decision

We will keep Architecture Decision Records in `docs/adr/`, one Markdown file per decision,
named `NNNN-kebab-title.md`, using [template.md](template.md) and listed in the index in
[README.md](README.md).

- A decision that is hard to reverse, or that touches interfaces, file formats, security,
  dependencies, platform scope or the project's process, MUST have an ADR.
- The ADR MUST be added or updated **in the same commit** as the change it records.
- Accepted ADRs are immutable except for their Status line and links. To change a decision,
  write a new ADR and mark the old one `Superseded by NNNN`.
- Spec Kit's per-feature `research.md` remains the working log for a feature; when a
  research item becomes a lasting decision it is promoted to an ADR and linked from there.
- This rule is part of the project constitution (Development Workflow).

## Consequences

Decisions are discoverable and reviewable in pull requests. Contributors pay a small cost
per significant change. The index must be kept current.

## Alternatives considered

- **Only `research.md` and the constitution**: decisions get buried in feature folders.
- **Install the community `adrkit` Spec Kit extension**: not vetted for install, and plain
  Markdown files need no tooling.
- **Wiki or issues**: not versioned with the code.
