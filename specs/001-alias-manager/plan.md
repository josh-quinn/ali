# Implementation Plan: Alias Manager (`ali`)

**Branch**: `001-alias-manager` | **Date**: 2026-11-08 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-alias-manager/spec.md`

## Summary

`ali` is a Go command-line tool that stores aliases in a hand-editable TOML file grouped by tool, renders a derived zsh file from it, and keeps the current shell in sync through a small marked block that setup adds to `.zshrc` once. Reads are pure file reads; the shell block never spawns `ali` at startup.

## Technical Context

**Language/Version**: Go 1.24+ (toolchain on dev machine: 1.27.3)

**Primary Dependencies**: `github.com/spf13/cobra`, `github.com/pelletier/go-toml/v2`

**Storage**: TOML file `$XDG_CONFIG_HOME/ali/aliases.toml` (default `~/.config/ali/`), derived `aliases.zsh` beside it

**Testing**: `go test` with table-driven unit tests; shell-level end-to-end tests run `zsh -f` in a temp `HOME`

**Target Platform**: macOS (arm64, amd64), zsh; Linux incidental

**Project Type**: single CLI project

**Performance Goals**: read commands < 10 ms; list/search 1,000 aliases < 50 ms; shell startup cost < 5 ms

**Constraints**: single static binary; no network; no writes outside `~/.config/ali/` except `init`/`import`

**Scale/Scope**: personal use, tens to low thousands of aliases

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | How the design satisfies it |
|-----------|--------|-----------------------------|
| I. Speed | PASS | Static binary; `.zshrc` block only `source`s a file, no `eval "$(ali ...)"` |
| II. Never harm shell | PASS | Only `init`/`import` touch `.zshrc`; backup + atomic rename + marker block; idempotent |
| III. Plain text store | PASS | TOML, one table per group; `aliases.zsh` always regenerated from it |
| IV. Test-first | PASS | Tasks order tests before implementation; E2E in temp HOME |
| V. Minimal deps | PASS | Only cobra and go-toml/v2 |
| VI. Scriptable CLI | PASS | `--json`, stderr errors, non-zero exits, `--force`, no prompts |

Post-design re-check: PASS (no violations; Complexity Tracking empty).

## Project Structure

### Documentation (this feature)

```text
specs/001-alias-manager/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli.md
├── checklists/requirements.md
└── tasks.md             # created by /speckit-tasks
```

### Source Code (repository root)

```text
cmd/ali/main.go                # wires cobra root, version via ldflags
internal/
├── alias/                     # Alias, Group types, name validation
├── group/                     # infer group from a command line
├── store/                     # load/save TOML, file lock, atomic write, backups
├── render/                    # aliases.zsh renderer, single-quote escaping, shell block text
├── shellconf/                 # insert/verify marked block in .zshrc (init), backup
├── importer/                  # parse `alias` lines from .zshrc
├── shadow/                    # detect names that shadow commands (PATH + builtins)
└── cli/                       # cobra commands: add, ls, groups, search, update, regroup, rm, init, import
tests/e2e/                     # zsh-driven end-to-end tests in temp HOME
.github/workflows/ci.yml
.goreleaser.yaml
README.md  LICENSE  CONTRIBUTING.md
```

**Structure Decision**: Single Go module with `internal/` packages, one concern per package, so each unit is testable without the CLI layer. `cli` is a thin adapter.

## Complexity Tracking

No constitution violations to justify.
