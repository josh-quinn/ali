# 0039. Distribute with a Homebrew cask built by GoReleaser

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Release preparation; supersedes ADR 0007

## Context

ADR 0007 chose a Homebrew tap formula via GoReleaser's `brews` section. `goreleaser check` (v2.20.0, first run during release preparation) reports `brews` as deprecated in favour of `homebrew_casks`, and an unvalidated release config would fail on the first tag. Unsigned, un-notarized binaries installed by a cask are quarantined by macOS Gatekeeper.

## Decision

We will publish with GoReleaser v2 to GitHub Releases (darwin and linux, arm64 and amd64, checksums) and a Homebrew cask in `josh-quinn/homebrew-tap`. Archives carry the binary, LICENSE, README and the generated man pages (sections 1 and 7), which the cask installs. A `postflight` step clears the quarantine attribute on macOS. The release repository is named in the config so `goreleaser check` works without a git remote. CI runs `goreleaser check` on every change.

## Consequences

One-line install once the tap exists (`brew install --cask josh-quinn/tap/ali`). Notarization remains a possible future improvement; until then the quarantine step is needed. The tap repository must be created and given a token before the first release.

## Alternatives considered

- **Keep `brews`**: deprecated; check fails.
- **Notarize the binary**: requires a paid Apple developer account; possible later.
