# 0022. Keep the config directory and its files private

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Security audit 2026-11-08

## Context

Aliases routinely embed secrets (`TOKEN=... curl`, database URLs). The store, generated file and lock were created `0644` inside a `0755` directory, readable by every local user.

## Decision

We will create the config directory `0700` when we create it, and write `aliases.toml`, `aliases.zsh` and the lock file `0600` on every update (which also tightens files from older versions). A directory that already exists is not chmod'ed, because `ALI_CONFIG_DIR` may point at a shared location. Backups of the user's `.zshrc` keep that file's own mode.

## Consequences

Other local users cannot read stored commands. Users who deliberately share the store (for example via a group-readable dotfiles checkout) must loosen permissions themselves; `ali` will tighten the files again on the next write.

## Alternatives considered

- **Leave defaults (umask)**: depends on the user's umask, often `022`.
- **chmod any existing directory**: could change the mode of `$HOME` if misconfigured.
