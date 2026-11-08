# 0042. Release from version tags with GoReleaser and build provenance

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: First release; complements ADR 0039

## Context

The first public release needs a repeatable process: binaries for macOS and Linux, a Homebrew cask, release notes, and a way for users to check that a download really came from this repository. ADR 0039 chose GoReleaser and a cask; how and from where releases run was still open.

## Decision

We will release only by pushing an annotated `vX.Y.Z` tag, following semantic versioning (pre-1.0: minor for features or incompatible changes, patch for fixes). The release workflow refuses tags whose commit is not on `main`, runs the full test suite, runs GoReleaser v2.20.0 (archives, checksums, release notes grouped by conventional-commit type, Homebrew cask), and attests build provenance for the archives and checksums with `actions/attest-build-provenance` (Sigstore, free for public repositories). The cask is pushed with a fine-grained token limited to the tap repository (`HOMEBREW_TAP_TOKEN`); without it the cask is skipped rather than failing the release. Only the release job gets `contents: write`, `id-token: write` and `attestations: write`. The steps are in RELEASING.md.

## Consequences

Releases are one command for a maintainer and verifiable for users (`gh attestation verify`). The tap token is a long-lived credential to rotate. Binaries are not code-signed or notarized by Apple; the cask clears the quarantine attribute instead (ADR 0039).

## Alternatives considered

- **Manual releases from a laptop**: not reproducible and no provenance.
- **Release on every merge to main**: too frequent for a CLI users install.
- **cosign keyless signing of checksums**: overlaps with GitHub attestations, which `gh` verifies directly.
