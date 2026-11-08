# 0040. Gate every change with free, open-source quality and security checks

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Release preparation; supersedes ADR 0026

## Context

The project is going public and every change must stay green on the platforms it supports. ADR 0026 set least-privilege CI with govulncheck and Dependabot but left actions pinned by tag. The tool also depends on real zsh behaviour that only end-to-end tests catch, and those must run on Linux as well as macOS.

## Decision

We will run, on every push to main and every pull request, using only free tools for public repositories:

- **Tests**: unit and end-to-end (real zsh driven through `zsh/zpty`) with the race detector and a coverage summary, on macOS and Linux, each with Go 1.24 (the `go.mod` minimum) and the latest stable Go.
- **Quality**: `gofmt`, `go vet`, golangci-lint v2 (standard linters plus revive, gocritic, errorlint, misspell, unconvert), the help/man completeness tests, actionlint on the workflows, `goreleaser check`.
- **Security**: gosec (inside golangci-lint), govulncheck, GitHub CodeQL (`security-and-quality` queries, weekly as well), OpenSSF Scorecard, dependency review on pull requests, Dependabot for Go modules and actions.

Every workflow has least-privilege `permissions`, checkouts do not persist credentials, and every third-party action is pinned to a full commit SHA with its version in a comment (Dependabot keeps them current). `make check` runs the local equivalent.

## Consequences

Mistakes surface before merge on both platforms and Go versions. CodeQL, Scorecard and dependency review require a public repository (or paid GitHub Advanced Security). CI takes a few minutes, dominated by the end-to-end suite.

## Alternatives considered

- **Paid SaaS (SonarCloud, Snyk)**: free tiers exist but add accounts and tokens; the chosen tools cover the same ground.
- **macOS only**: misses Linux zsh differences.
