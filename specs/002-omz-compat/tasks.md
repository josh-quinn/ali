---

description: "Task list for Oh My Zsh compatibility"
---

# Tasks: Oh My Zsh compatibility

**Input**: `/specs/002-omz-compat/` (plan.md, spec.md, research.md, data-model.md, contracts/cli.md)
**Tests**: required, written first (Constitution IV). Each task is one green, conventional commit with its tests.

## Phase 1: Setup
- [ ] T001 Mark ADRs 0027-0029 Accepted (status line only) once the owner approves this plan

## Phase 2: Foundational (blocks US1, US2, US4)
- [ ] T002 Tests then `internal/shellconf`: locate block markers, `Install` replaces an outdated block in place with backup; error on unbalanced/duplicate markers; bytes outside the block unchanged (research R4)
- [ ] T003 Tests then `internal/render`: block v2 text with version line; wrapper still hijack-safe (ADR 0021)

## Phase 3: US1 Warn on overriding shell aliases/functions (P1) 🎯 MVP
- [ ] T004 [P] [US1] Tests then `internal/shelldef`: parse `ALI_SHELL_DEF` (alias/function/empty, malformed → ignored)
- [ ] T005 [US1] Tests then `internal/render`: wrapper sets `ALI_SHELL_DEF` for `add` only, excluding ali-owned names
- [ ] T006 [US1] Tests then `internal/cli/add.go`: print override warning (FR-001..003)
- [ ] T007 [US1] E2E `tests/e2e/omz_test.go`: plugin-like alias and function defined before the block → warnings, aliases created

## Phase 4: US2 Restore originals on removal (P1)
- [ ] T008 [US2] Tests then `internal/render`: `_ali_prev` record/restore logic in `aliases.zsh` (research R2, incl. `zsh/parameter` check)
- [ ] T009 [US2] E2E: add over `gp`, rm → `git push` restored in the same shell; update keeps the original remembered

## Phase 5: US4 Upgrade the block (P2)
- [ ] T010 [US4] Tests then `internal/cli/init.go`: "updated" outcome and messages
- [ ] T011 [US4] E2E: v1 block + surrounding content → upgraded, rest byte-identical, second run no-op

## Phase 6: US3 Doctor (P2)
- [ ] T012 [P] [US3] Tests then `internal/doctor`: checks and result model (data-model.md)
- [ ] T013 [US3] Tests then `internal/render`: wrapper pipes shell state to `ali doctor --shell-state -`
- [ ] T014 [US3] Tests then `internal/cli/doctor.go`: output and exit codes (contracts/cli.md)
- [ ] T015 [US3] E2E: each misconfiguration flagged; healthy setup passes; outside-wrapper run lists skipped checks

## Phase 7: Polish
- [ ] T016 [P] README: "Using ali with Oh My Zsh" section; update Platform support/limitations text about shadow warnings
- [ ] T017 [P] CHANGELOG entry
- [ ] T018 Run quickstart.md; verify SC-001..SC-005 (startup delta < 1 ms)

## Dependencies
Phase 2 → US1 → US2; US4 and US3 depend only on Phase 2 (US3's live check also needs T005's wrapper pattern). MVP = Phases 2-3.
