# 0044. Accept one tag-pinned reference for the SLSA provenance generator

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Scorecard run of 2026-11-09; corrects ADR 0043

## Context

ADR 0043 stated that OpenSSF Scorecard exempts `slsa-framework/slsa-github-generator` from its pinning rule. It does not: Scorecard v5.6.0 reports `third-party GitHubAction not pinned by hash: .github/workflows/release.yml`, lowering Pinned-Dependencies from 10 to 9. The generator cannot simply be pinned by commit SHA, because slsa-verifier verifies the provenance builder by its release tag; a SHA reference makes the provenance unverifiable.

## Decision

We will keep referencing the SLSA generator by its release tag (currently v2.1.0), as its maintainers require, and accept the one-point Pinned-Dependencies deduction. Every other action stays pinned by full commit SHA, and Dependabot keeps the generator tag current.

## Consequences

Releases keep SLSA level 3 provenance that slsa-verifier can check (verified for v0.1.2), worth up to 10 points on Signed-Releases, at the cost of 1 point on Pinned-Dependencies.

## Alternatives considered

- **Pin the generator by SHA**: breaks provenance verification.
- **Replace the SLSA generator with a Sigstore bundle from actions/attest-build-provenance uploaded as `.sigstore.json`**: fully pinned, but Scorecard rates a signature below provenance, and the result is not verifiable with slsa-verifier.
