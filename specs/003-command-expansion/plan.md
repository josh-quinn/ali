# Implementation Plan: Command expansion

**Branch**: `003-command-expansion` | **Date**: 2026-11-08 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/003-command-expansion/spec.md` (depends on feature 002 for the in-place block upgrade)

## Summary

One expansion engine, written in Go and unit-tested, turns a command line into the command to run: it finds the tool, expands a leading run of sub-aliases, resolves `{{name}}` placeholders from `@` words, the environment and defaults, and inserts values as safely quoted words. zsh only provides thin glue generated into `aliases.zsh`: line-editor widgets on Space and Enter (visible mode) or small functions (run-time mode). The glue calls the binary only when a pre-generated table says the line can involve an ali expansion, so ordinary typing never starts a process. Environments are chosen per terminal with an exported variable set by the `ali()` wrapper, or globally via a machine-local file, and shown by a prompt hook that reads shell variables only.

## Technical Context

**Language/Version**: Go 1.24+; generated zsh 5.x (ZLE widgets, `add-zsh-hook`)

**Primary Dependencies**: unchanged (cobra, go-toml/v2). No pty library: interactive tests use zsh's own `zsh/zpty` module.

**Storage**: `aliases.toml` gains `_subs`, `_shared`, `_vars`, `_envs` tables (format version 2 written only when used); new machine-local `settings.toml` and `global-env` in the config dir

**Testing**: table-driven engine tests (tokenizer, expansion, precedence, hostile values); render tests for generated zsh; e2e through a real line editor driven by `zsh/zpty`

**Target Platform**: macOS + zsh

**Project Type**: CLI + generated shell integration

**Performance Goals**: zero processes for lines with no ali expansion; < 15 ms per expanding keypress (one binary run, measured ~5-8 ms today); prompt hook spawns nothing

**Constraints**: hijack-safe generated code (ADR 0021); coexist with plugins that wrap Space/Enter; values never evaluated (Constitution VII)

**Scale/Scope**: hundreds of sub-aliases, tens of variables and environments

## Constitution Check

| Principle | Status | Note |
|-----------|--------|------|
| I. Speed | PASS | Gating table in zsh; one process only on expanding keypresses; no startup process |
| II. Never harm shell | PASS | Block change ships via feature 002's in-place upgrade; key bindings chain to existing widgets |
| III. Plain text store | PASS | New TOML tables; settings in a separate machine-local file |
| IV. Test-first | PASS | Engine fully unit-tested; ZLE behaviour e2e-tested with `zsh/zpty` |
| V. Scope (v2.0.0) | PASS | Exactly the amended scope: sub-aliases, placeholders, environments |
| VI. Scriptable | PASS | `ali expand` prints without running; `--json` on listings; no prompts except protected environments, which are the explicit exception |
| VII. Show what runs | PASS | Visible mode default; history records expanded text; values quoted as data; prompt indicator |

**ADRs**: proposes 0030-0037 (see [docs/adr](../../docs/adr/README.md)).

## Project Structure

### Documentation (this feature)

```text
specs/003-command-expansion/
├── plan.md  research.md  data-model.md  quickstart.md
├── contracts/cli.md
├── checklists/requirements.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/model/       # NEW: Sub, Variable, Environment types + validation
internal/store/       # v2 tables, versioning, referential checks
internal/settings/    # NEW: settings.toml + global-env (machine-local)
internal/expand/      # NEW: tokenizer, tool detection, sub expansion, placeholders, @ parsing, quoting
internal/render/      # gating tables, ZLE widgets, run-time functions, env wrapper, prompt hook, completion
internal/cli/         # sub.go, var.go, env.go, config.go, expand.go, show.go
tests/e2e/zle/        # zpty-driven interactive tests
```

**Structure Decision**: the engine is a pure package (`expand`) with no I/O so every rule in the spec is a unit test; `render` owns all zsh text.

## Complexity Tracking

| Item | Why needed | Simpler alternative rejected because |
|------|------------|--------------------------------------|
| ZLE widgets on Space and Enter | Visible expansion (owner decision 1) | Plain aliases cannot expand words after the first |
| `zsh/zpty` test driver | ZLE is inactive without a terminal | A Go pty library would add a dependency (Constitution V) |
