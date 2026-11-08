# 0037. Store format v2 only when used, with machine-local settings kept separate

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: specs/003-command-expansion research R12, data-model

## Context

New tables must live in the human-editable store (Constitution III), and users sync that file across machines, some possibly running an older ali. Preferences such as expansion mode and the global environment are per machine.

## Decision

We will add `_subs`, `_shared`, `_vars` and `_envs` tables to `aliases.toml`, writing `version = 2` only when any exist (else `version = 1`), so older binaries refuse rather than corrupt a v2 store. Machine-local preferences go in `settings.toml` and the global environment in `global-env`, both in the config directory with private permissions and not meant to be synced.

## Consequences

Existing stores are unaffected until a new feature is used. Users syncing the config directory wholesale should exclude the two machine-local files (documented).

## Alternatives considered

- **Separate store files per concept**: harder to sync and edit together.
- **Settings inside `aliases.toml`**: would sync machine preferences.
