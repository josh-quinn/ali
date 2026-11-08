---

description: "Task list for the Alias Manager (ali)"
---

# Tasks: Alias Manager (`ali`)

**Input**: Design documents from `/specs/001-alias-manager/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md
**Tests**: Included and written first (Constitution IV, Test-First).

## Format: `[ID] [P?] [Story] Description`
- **[P]**: parallel (different files, no dependencies)

## Phase 1: Setup

- [x] T001 `go mod init github.com/josh-quinn/ali`; add cobra and go-toml/v2; create directory layout from plan.md
- [x] T002 [P] Add `.gitignore`, `LICENSE` (MIT), `.editorconfig`, `Makefile` (build, test, vet, fmt-check)
- [x] T003 [P] Add `.github/workflows/ci.yml` (macos-latest: gofmt check, vet, test)

## Phase 2: Foundational (blocks all stories)

- [x] T004 [P] Tests then `internal/alias/alias.go`: types and name/group validation (data-model.md rules)
- [x] T005 [P] Tests then `internal/group/group.go`: group inference (sudo/env/VAR=x/path/empty→misc, research R6)
- [x] T006 [P] Tests then `internal/render/quote.go` + `render.go`: single-quote escaping and `aliases.zsh` rendering with `_ali_names` unalias header (R2, R3); include `'`, `$`, backtick, newline, unicode cases
- [x] T007 Tests then `internal/store/store.go`: load/save TOML, duplicate-name detection, atomic write (temp+fsync+rename), flock; config dir resolution (`ALI_CONFIG_DIR`, `XDG_CONFIG_HOME`)
- [x] T008 `internal/cli/root.go` + `cmd/ali/main.go`: cobra root, `--version` via ldflags, error→stderr/exit codes, shared "load store, mutate, save, render" helper

**Checkpoint**: store round-trips and renders; CLI skeleton builds.

## Phase 3: User Story 1 - Create an alias on the fly (P1) 🎯 MVP

**Independent test**: `ali add gp 'git pull'` then `zsh -i -c gp`-style check in temp HOME.

- [x] T009 [P] [US1] Tests for `internal/shadow/shadow.go` (builtins + PATH) then implement
- [x] T010 [P] [US1] CLI tests for `add` (new, exists, `--force`, `-g`, invalid name, special chars) in `internal/cli/add_test.go`
- [x] T011 [US1] Implement `internal/cli/add.go` per contracts/cli.md
- [x] T012 [P] [US1] Tests then `internal/shellconf/shellconf.go`: insert marked block into `.zshrc` with timestamped backup, idempotent, honours `ZDOTDIR`
- [x] T013 [US1] Implement `internal/cli/init.go` (`ali init zsh [--dry-run]`)
- [x] T014 [US1] E2E `tests/e2e/add_test.go`: init, add, new `zsh -i` and same-shell wrapper both see the alias; real HOME untouched

## Phase 4: User Story 2 - See and search (P1)

**Independent test**: seeded store → `ls`, `ls helm`, `groups`, `search pods`, `--json`.

- [x] T015 [P] [US2] Tests then `internal/cli/ls.go` (grouped, aligned, group filter, unknown group error, `--json`)
- [x] T016 [P] [US2] Tests then `internal/cli/groups.go`
- [x] T017 [P] [US2] Tests then `internal/cli/search.go` (case-insensitive over name/command/group, no match → exit 1)
- [x] T018 [US2] Benchmark test: 1,000 aliases list/search < 50 ms (SC-003)

## Phase 5: User Story 3 - Update and delete (P2)

**Independent test**: update changes behaviour in same shell; rm makes alias undefined.

- [x] T019 [P] [US3] Tests then `internal/cli/update.go` (pinned vs inferred group, missing name error)
- [x] T020 [P] [US3] Tests then `internal/cli/regroup.go`
- [x] T021 [P] [US3] Tests then `internal/cli/rm.go` (multiple names, partial failure, drop empty group)
- [x] T022 [US3] E2E `tests/e2e/update_rm_test.go`: removed alias is unaliased in the live shell via `_ali_names`

## Phase 6: User Story 4 - Import existing aliases (P2)

**Independent test**: sandbox `.zshrc` with mixed alias lines → import summary + backup, originals intact.

- [x] T023 [P] [US4] Tests then `internal/importer/importer.go` (single/double/bare quotes, `alias -g` rejected, unparseable lines reported)
- [x] T024 [US4] Tests then `internal/cli/import.go` (`--file`, `--dry-run`, conflicts skipped, backup, lists originals to delete)

## Phase 7: Polish & public release

- [x] T025 [P] `README.md`: problem statement, install (brew + go install), quick start, command reference, FAQ (why a wrapper function)
- [x] T026 [P] `CONTRIBUTING.md`, issue templates, `CHANGELOG.md`
- [x] T027 [P] `.goreleaser.yaml` (darwin arm64/amd64, checksums, Homebrew tap); `goreleaser check` passes with GoReleaser v2.20.0 (feature 003 release prep, ADR 0039)
- [x] T028 Run quickstart.md end-to-end in a sandbox; verify SC-001..SC-007 (incl. SC-004: time `zsh -i -c exit` with and without the block, delta < 5 ms); `go vet`, `gofmt -l .`, `go test ./...`
- [ ] T029 Commit; ask the user before creating the public GitHub repo or pushing

## Dependencies
- Phase 1 → Phase 2 → stories. US1 needs T004-T008 and delivers the MVP; US2/US3/US4 depend only on Phase 2 (+ T013 for E2E). Parallel marks `[P]` are safe within a phase.

## Implementation strategy
MVP = Phases 1-3 (add + setup). Then US2, US3, US4 incrementally, committing after each task group with tests green.
