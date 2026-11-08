# ali Constitution

## Core Principles

### I. Speed Is a Feature
`ali` runs many times a day and sits on the critical path of the shell. Startup MUST stay
under 10 ms for read operations, and the shell hook sourced from `.zshrc` MUST NOT spawn
the binary on shell startup. The product ships as a single static binary with no runtime.
Rationale: a tool meant to preserve flow cannot itself add latency.

### II. Never Harm the User's Shell
`ali` MUST NOT modify `~/.zshrc` or any user-owned file except through explicit, opt-in
commands (`ali init`, `ali import`). Every such write MUST be preceded by a timestamped
backup, MUST be idempotent, and MUST be atomic (write temp file, then rename). Normal
add/update/remove commands touch only `ali`'s own files under `~/.config/ali/`. A corrupt
or hand-broken store MUST produce a clear error, never data loss or a broken shell.

### III. Plain Text, Hand-Editable Store
Aliases are persisted in a human-readable TOML file, one table per group. Users MUST be
able to edit, diff, version-control and sync it without `ali`. The generated zsh file is a
derived artifact and MUST be fully reproducible from the store.

### IV. Test-First (NON-NEGOTIABLE)
TDD is mandatory: write a failing test, then the implementation, then refactor. Shell
quoting and escaping (single quotes, `$`, backticks, newlines, unicode), grouping
inference, atomic writes and import parsing MUST each have explicit edge-case tests.
End-to-end tests MUST run against a throwaway `HOME` and never touch the real one.

### V. Simplicity and Minimal Dependencies
Scope for v1 is aliases only, zsh only, macOS first. Features beyond that (functions,
env vars, other shells, TUI) require an amendment-level justification. Third-party
dependencies are limited to `spf13/cobra` and `pelletier/go-toml/v2`; anything else MUST
be justified in the plan's complexity tracking.

### VI. Predictable, Scriptable CLI
Output for humans goes to stdout, diagnostics and errors to stderr, with non-zero exit
codes on failure. Listing commands MUST offer a stable machine-readable form (`--json`).
Destructive or overwriting actions require an explicit flag (`--force`) rather than a
prompt, so the tool stays non-interactive and scriptable.

## Public Open-Source Standards

The repository MUST be release-ready for GitHub: MIT license, README with install and
quick-start, CONTRIBUTING notes, CI running `go vet`, `gofmt` check and `go test` on macOS,
and GoReleaser producing signed-checksum release artifacts plus a Homebrew tap formula.
Releases follow semantic versioning. User-visible behaviour is documented in the README
and the spec-kit artifacts under `specs/`.

## Development Workflow

Work flows through spec-kit: constitution, specify, clarify, plan, tasks, analyze,
implement. Each change set is small, committed with a conventional-commit message, and
leaves `go test ./...`, `go vet ./...` and `gofmt -l .` clean. The plan's Constitution
Check MUST be re-run before implementation starts.

## Governance

This constitution supersedes other practices. Amendments require a written rationale in
the commit message and a version bump: MAJOR for removing or redefining a principle,
MINOR for adding a principle or materially expanding guidance, PATCH for wording fixes.
Every plan and review MUST verify compliance; added complexity MUST be justified in the
plan's complexity tracking table.

**Version**: 1.0.0 | **Ratified**: 2026-11-08 | **Last Amended**: 2026-11-08
