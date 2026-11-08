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

Every command MUST have a one-line summary, a long description and at least one example
in its `--help`, and syntax that is typed in the shell (not as an `ali` subcommand) MUST be
explained in a help topic (`ali help <topic>`). Man pages MUST be generated from the same
command definitions (one page per command plus topic pages) and shipped with releases. A
test MUST fail when a command is missing any of these.

## Development Workflow

Work flows through spec-kit: constitution, specify, clarify, plan, tasks, analyze,
implement. The plan's Constitution Check MUST be re-run before implementation starts.

### Commits (project rule)
- Every commit MUST be small and single-purpose, and MUST leave `make fmt-check vet test`
  green, so any commit can be checked out and tested on its own.
- Commit messages MUST follow [Conventional Commits](https://www.conventionalcommits.org/):
  `type(scope): imperative subject`, with types `feat`, `fix`, `docs`, `test`, `build`, `ci`,
  `refactor`, `perf`, `chore`. Breaking changes use `!` or a `BREAKING CHANGE:` footer.
- Tests are committed together with the code they cover (never a red commit followed by a
  fix). A bug fix MUST include a regression test that fails without the fix.
- Explain *why* in the body whenever the subject alone would not make it obvious.
- Unrelated changes MUST NOT share a commit. History is only rewritten before it is pushed.

### Architecture Decision Records (project rule)
- Decisions that are hard to reverse, or that touch interfaces, file formats, security,
  dependencies, platform scope or process, MUST be recorded as an ADR in `docs/adr/`
  (see ADR 0001), in the same commit as the change.
- Accepted ADRs are not edited except for status and links; a changed decision gets a new
  ADR that supersedes the old one. The index in `docs/adr/README.md` MUST stay current.
- A plan's Constitution Check MUST list the ADRs it adds or supersedes.

### Security (project rule)
- Anything that writes outside `~/.config/ali/`, parses untrusted text, or emits shell code
  MUST be reviewed for injection, path and permission issues, with a regression test for
  every issue found.

## Governance

This constitution supersedes other practices. Amendments require a written rationale in
the commit message and a version bump: MAJOR for removing or redefining a principle,
MINOR for adding a principle or materially expanding guidance, PATCH for wording fixes.
Every plan and review MUST verify compliance; added complexity MUST be justified in the
plan's complexity tracking table.

**Version**: 1.2.0 | **Ratified**: 2026-11-08 | **Last Amended**: 2026-11-08
