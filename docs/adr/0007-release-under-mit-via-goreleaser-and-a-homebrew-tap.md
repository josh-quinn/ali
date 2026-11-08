# 0007. Release under MIT via GoReleaser and a Homebrew tap

- **Status**: Superseded by [0039](0039-distribute-with-a-homebrew-cask-built-by-goreleaser.md)
- **Date**: 2026-11-08
- **Spec**: Constitution, Public Open-Source Standards

## Context

The tool is intended to be public and installable with one command.

## Decision

We will license under MIT and release with GoReleaser (darwin and linux, arm64 and amd64, checksums) and a Homebrew tap `josh-quinn/homebrew-tap`. The GoReleaser config has not yet been validated because `goreleaser` is not installed on the author's machine.

## Consequences

One-line install for macOS users once the tap repo exists. The README says Homebrew is available only after the first release.

## Alternatives considered

- **GPL/Apache-2.0**: MIT is simplest for a small utility.
- **`go install` only**: requires a Go toolchain.
