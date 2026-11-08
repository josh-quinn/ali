# 0023. Write through symlinks instead of replacing them

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Security audit 2026-11-08

## Context

Atomic replacement (`rename` over the path) swaps a symlink for a regular file. Many users keep `.zshrc` (and would keep `aliases.toml`) as symlinks into a dotfiles repo managed by stow, chezmoi or yadm. `ali init` silently detached `.zshrc` from the repo, and the next `ali add` would have done the same to the store, so those users' edits stopped reaching version control. This is also an integrity bug: the write did not land where the user's config actually lives.

## Decision

We will resolve symlinks (up to 16 hops, relative links resolved against the link's directory, dangling links resolved to their target) inside `store.WriteAtomic` before creating the temp file, so the temp file and the rename happen in the target's directory. This applies to every file `ali` writes, including `.zshrc` and backups of it.

## Consequences

Symlinked dotfiles keep working. The temp file must live on the target's filesystem, which resolving first guarantees. A symlink pointing somewhere unexpected is followed; that is the user's own configuration, so no extra trust boundary is crossed.

## Alternatives considered

- **Refuse to write symlinks**: breaks the common dotfiles workflow.
- **Write non-atomically through the link**: risks torn files.
