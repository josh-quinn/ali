---

description: "Task list for command expansion"
---

# Tasks: Command expansion

**Input**: `/specs/003-command-expansion/` (plan, spec, research, data-model, contracts)
**Prerequisite**: feature 002 merged (in-place block upgrade, ADR 0029).
**Tests**: required and written first; each task is one green conventional commit including its tests and any ADR status change.

## Phase 1: Setup
- [x] T001 Mark ADRs 0030-0037 Accepted once the owner approves this plan
- [x] T002 [P] `tests/e2e/zle_helper_test.go`: zpty driver helper (send keys, read line, tolerate bracketed-paste escapes) with a self-test (research R10)

## Phase 2: Foundational (blocks all stories)
- [x] T003 [P] Tests then `internal/model`: Sub, Variable, Environment types and validation rules (data-model.md)
- [x] T004 Tests then `internal/store`: `_subs`, `_shared`, `_vars`, `_envs` load/save; version 1/2 selection; referential checks (placeholders reference variables; env sets existing variables)
- [x] T005 [P] Tests then `internal/settings`: `settings.toml` and `global-env`, private permissions (ADR 0022)
- [x] T006 [P] Tests then `internal/expand` tokenizer: raw spans, quoted/escaped flags, command separators (research R6)
- [x] T007 Tests then `internal/expand` value quoting (research R7) incl. hostile-value table (SC-004)

## Phase 3: US1 Tool-scoped sub-aliases, visible expansion (P1) 🎯 MVP
- [x] T008 [US1] Tests then `internal/expand`: tool detection through aliases and wrappers; sub-alias expansion anywhere after the tool (`k get p`, `k logs x p`); tool-over-shared; `\word`; bare `--` ends expansion; per-command scoping across `;`/`|`/`&&`; no recursion
- [x] T009 [P] [US1] Tests then `internal/cli/sub.go` (add/update/rm/ls, `--shared`, `--json`)
- [x] T010 [US1] Tests then `internal/cli/expand.go` (`ali expand '<line>'` and `--zle` protocol, research R4)
- [x] T011 [US1] Tests then `internal/render`: `_ali_trigger` and sub-name gate tables; Space and Enter widgets chaining to previous widgets; paste skipped (research R2, R3)
- [x] T012 [US1] E2E (zpty): `k g p ` shows `kubectl get pods `; `k get p ` expands `p`; `k logs \p` stays literal; Enter without space expands; history text; `kubectl g p`; plain `ls` spawns no process

## Phase 4: US2 Placeholders and @ values (P1)
- [ ] T013 [US2] Tests then `internal/expand`: placeholder parsing (defined variables only), `@key`, `@var=value`, ambiguity error, unmatched `@` words left alone, same value for repeated placeholders
- [ ] T014 [P] [US2] Tests then `internal/cli/var.go` (add/set/unset/default/rm/ls/show, references)
- [ ] T015 [US2] Tests then save-time validation in `add`/`update`/`sub`: unknown-variable placeholders literal with notice; placeholders inside quotes rejected
- [ ] T016 [US2] Tests then `internal/render`: aliases containing placeholders rendered so they never run unexpanded (function fallback)
- [ ] T017 [US2] E2E (zpty): `k g p n @sys`, default, `@ns=` literal, unresolved stays in editor with message, hostile value as one argument

## Phase 5: US3 Environments (P2)
- [ ] T018 [US3] Tests then `internal/expand`: precedence chain (FR-009), `@@env`
- [ ] T019 [P] [US3] Tests then `internal/cli/env.go` (add/set/unset/protect/rm/ls/show/use/off/current)
- [ ] T020 [US3] Tests then `internal/render`: wrapper handles `env use|off` (export/unset `ALI_ENV`), precmd prompt hook (research R8, R9); the hook must validate the environment name itself before putting it in RPROMPT (global-env and ALI_ENV are user-editable, and prompt_subst may be on), with a hostile-name test
- [ ] T021 [US3] Tests then protected-environment confirm path in widget (`confirm` status → second Enter)
- [ ] T022 [US3] E2E (zpty, two sessions): per-terminal isolation, `--global`, prompt label and colour, protected double-Enter

## Phase 6: US4 Run-time mode (P2)
- [ ] T023 [US4] Tests then `internal/cli/config.go` and regeneration on `expand` change
- [ ] T024 [US4] Tests then `internal/render`: run-time functions with ali-quoted `eval` (research R5); `/dev/tty` confirm for protected
- [ ] T025 [US4] E2E: line unchanged while typing, same command runs; tool-first line not expanded

## Phase 7: US5 Discover and complete (P3)
- [ ] T026 [P] [US5] Tests then `internal/cli/show.go`
- [ ] T027 [US5] Tests then generated completion for ali commands and Tab widget for `@` values and sub-aliases, falling through otherwise (research R11)
- [ ] T028 [US5] E2E (zpty): `@<TAB>` offers `one sys`; `k <TAB>` offers kubectl sub-aliases

## Documentation (applies to every CLI task above)
Each CLI task (T009, T010, T014, T019, T023, T026) includes Short, Long and Example help, enforced by the help completeness test from feature 002 (T019 there); man pages regenerate automatically.
- [ ] T033 [P] Help topics `expansion`, `placeholders`, `environments` with examples; man pages generated for them (FR-019)

## Phase 8: Polish
- [ ] T029 [P] README: sub-aliases, placeholders, environments, settings; update Platform support/limitations
- [ ] T030 [P] CHANGELOG
- [ ] T031 Performance: SC-003 measurements recorded in research.md
- [ ] T032 Manual compatibility pass with zsh-autosuggestions and zsh-syntax-highlighting; record result (and any limitation) in README

## Dependencies
Phase 2 → US1 → US2 → US3; US4 needs US2; US5 needs US1-US2. MVP = Phases 1-3 (sub-aliases with visible expansion). Within phases, `[P]` tasks touch different files.
