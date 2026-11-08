# 0026. Harden CI and keep dependencies patched

- **Status**: Superseded by [0040](0040-gate-every-change-with-free-open-source-quality-and-security-checks.md)
- **Date**: 2026-11-08
- **Spec**: Security audit 2026-11-08

## Context

CI runs third-party actions and, later, will run the release pipeline. By default a workflow token can be broad, checkout persists credentials in the git config, and nothing watches the two Go dependencies or the actions for vulnerabilities. `govulncheck` was run on 2026-11-08 and reported no known vulnerabilities.

## Decision

We will declare `permissions: contents: read` in the CI workflow, disable credential persistence on checkout, run `govulncheck` (pinned to v1.9.0) on every push, and enable Dependabot for Go modules and GitHub Actions on a weekly schedule. Actions are currently referenced by major-version tag; pinning them to commit SHAs is a recommended follow-up that needs network access to resolve and was not done here.

## Consequences

A compromised or buggy step cannot push or write. Vulnerable dependencies surface automatically. Tag-pinned actions remain a residual supply-chain risk until pinned by SHA. Release workflows (GoReleaser) will need a narrowly scoped `contents: write` job when added.

## Alternatives considered

- **Leave defaults**: broader token than needed.
- **Pin actions by SHA now**: could not verify SHAs offline.
