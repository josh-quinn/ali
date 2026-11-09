# 0043. Meet OpenSSF Scorecard expectations where a solo project can

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Owner request; Scorecard 6.1/10 on 2026-11-08

## Context

The first OpenSSF Scorecard run scored 6.1. Several checks are fixable (Fuzzing 0, Security-Policy 0, Branch-Protection 0, Signed-Releases with no releases); others cannot be fixed by a single maintainer or only improve with time (Code-Review needs approvals from a second person, Maintained needs 90 days of history, Contributors needs several organizations, CI-Tests and SAST are lowered by early direct pushes to main, CII-Best-Practices needs the owner to register the project).

## Decision

We will:

- **Fuzzing**: keep Go native fuzz tests for the tokenizer, quoting, the expansion engine, the importer, the store parser and group inference, each asserting a property; CI runs every target briefly on each change (`make fuzz`).
- **Security policy**: keep `SECURITY.md` with private reporting through GitHub private vulnerability reporting (enabled).
- **Signed releases**: add signed SLSA level 3 provenance (`ali.intoto.jsonl`) to every release with `slsa-framework/slsa-github-generator` (referenced by tag, which it requires and Scorecard exempts), alongside the GitHub attestations of ADR 0042, because Scorecard only inspects release assets.
- **Branch protection**: protect `main` with a repository ruleset: no deletion, no force-push, changes only through pull requests, and the CI test, lint, secrets, fuzz and CodeQL checks required to pass.
- Not pursue checks that need other people or time, and document what the owner can do (register at bestpractices.dev; add a second reviewer when the project has one).

## Consequences

Scorecard rises to the level a well-run solo project can reach; the remaining gaps are explained rather than gamed (for example, no fake second approver). Requiring pull requests and passing checks means even the owner merges through PRs, which the project already does.

## Alternatives considered

- **Require one approving review**: a solo maintainer cannot approve their own pull requests, so every merge would need an admin bypass, which Scorecard also penalizes.
- **cosign-signed checksums instead of SLSA provenance**: scores slightly lower and duplicates the attestations.
