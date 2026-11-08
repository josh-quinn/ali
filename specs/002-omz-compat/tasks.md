---

description: "Task list for Oh My Zsh compatibility"
---

# Tasks: Oh My Zsh compatibility

**Input**: `/specs/002-omz-compat/` (plan.md, spec.md, research.md, data-model.md, contracts/cli.md)
**Tests**: required, written first (Constitution IV). Each task is one green, conventional commit with its tests.

## Phase 1: Setup
- [x] T001 Mark ADRs 0027-0029 Accepted (status line only) once the owner approves this plan
- [x] T019 Tests then help completeness: a test walks the cobra tree and fails if any command lacks Short, Long or Example; fill in Long/Example for every existing command (add, ls, groups, search, update, regroup, rm, init, import) (FR-008)
- [x] T020 Tests then `ali help files` topic (store, generated file, lock, backups, settings) (FR-008)
- [x] T021 `internal/tools/gendocs` + `make man` generating `man/*.1` (ali.1, per command, topics); GoReleaser `before` hook and archive files; Homebrew formula installs pages; test that generation succeeds and every command has a page (ADR 0038, SC-006)

## Phase 2: Foundational (blocks US1, US2, US4)
- [x] T002 Tests then `internal/shellconf`: locate block markers, `Install` replaces an outdated block in place with backup; error on unbalanced/duplicate markers; bytes outside the block unchanged (research R4)
- [x] T003 Tests then `internal/render`: block v2 text with version line; wrapper still hijack-safe (ADR 0021)

## Phase 3: US1 Warn on overriding shell aliases/functions (P1) 🎯 MVP
- [x] T004 [P] [US1] Tests then `internal/shelldef`: parse `ALI_SHELL_DEF` (alias/function/empty, malformed → ignored)
- [x] T005 [US1] Tests then `internal/render`: wrapper sets `ALI_SHELL_DEF` for `add` only, excluding ali-owned names
- [x] T006 [US1] Tests then `internal/cli/add.go`: print override warning (FR-001..003)
- [x] T007 [US1] E2E `tests/e2e/omz_test.go`: plugin-like alias and function defined before the block → warnings, aliases created

## Phase 4: US2 Restore originals on removal (P1)
- [x] T008 [US2] Tests then `internal/render`: `_ali_prev` record/restore logic in `aliases.zsh` (research R2, incl. `zsh/parameter` check)
- [x] T009 [US2] E2E: add over `gp`, rm → `git push` restored in the same shell; update keeps the original remembered

## Phase 5: US4 Upgrade the block (P2)
- [x] T010 [US4] Tests then `internal/cli/init.go`: "updated" outcome and messages
- [x] T011 [US4] E2E: v1 block + surrounding content → upgraded, rest byte-identical, second run no-op

## Phase 6: US3 Doctor (P2)
- [x] T012 [P] [US3] Tests then `internal/doctor`: checks and result model (data-model.md)
- [ ] T013 [US3] Tests then `internal/render`: wrapper pipes shell state to `ali doctor --shell-state -`
- [ ] T014 [US3] Tests then `internal/cli/doctor.go`: output and exit codes (contracts/cli.md), with Long and Example help
- [ ] T015 [US3] E2E: each misconfiguration flagged; healthy setup passes; outside-wrapper run lists skipped checks

## Phase 7: Polish
- [ ] T016 [P] README: "Using ali with Oh My Zsh" section; update Platform support/limitations text about shadow warnings; document `ali help` and `man ali`
- [ ] T017 [P] CHANGELOG entry
- [ ] T018 Run quickstart.md; verify SC-001..SC-005 (startup delta < 1 ms)

## Dependencies
Phase 1 docs tasks (T019-T021) come first so every later command is written to the help standard. Phase 2 → US1 → US2; US4 and US3 depend only on Phase 2 (US3's live check also needs T005's wrapper pattern). MVP = Phases 2-3.
