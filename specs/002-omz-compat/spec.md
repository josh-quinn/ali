# Feature Specification: Oh My Zsh compatibility

**Feature Branch**: `002-omz-compat`

**Created**: 2026-11-08

**Status**: Draft (awaiting owner approval)

**Input**: User description: "How does this work with Oh My Zsh? It should work with it." Follow-up to the review finding that shadow warnings cannot see aliases defined by Oh My Zsh (e.g. the git plugin's `gp='git push'`).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Warn when overriding an alias or function already in my shell (Priority: P1)

I use Oh My Zsh with the git plugin. When I run `ali add gp 'git pull'`, I want to be told that `gp` is already `git push` in my shell, so I don't lose a habit by accident.

**Why this priority**: Silent overrides of plugin aliases are the main surprise for Oh My Zsh users.

**Independent Test**: In a sandbox shell where `alias gp='git push'` and a function `mk` are defined before the ali block, `ali add gp 'git pull'` and `ali add mk 'make'` each print a warning naming the existing definition, and both aliases are still created.

**Acceptance Scenarios**:

1. **Given** the shell has `alias gp='git push'`, **When** I run `ali add gp 'git pull'`, **Then** stderr shows `warning: gp is already an alias in this shell (git push); ali's alias will take precedence` and the alias is created.
2. **Given** the shell has a function `mk`, **When** I run `ali add mk make`, **Then** I am warned that `mk` is a shell function.
3. **Given** `gp` is already an ali alias, **When** I run `ali update gp ...`, **Then** no override warning is shown.
4. **Given** the binary is run directly (not through the shell wrapper, e.g. from a script), **When** I add an alias, **Then** the existing builtin/PATH warning still works and nothing fails.

---

### User Story 2 - Removing my override brings the original back (Priority: P1)

If I add `gp` over Oh My Zsh's `gp` and later `ali rm gp`, I expect `gp` to be `git push` again immediately, not undefined until I open a new terminal.

**Why this priority**: Without it, removal leaves the live shell inconsistent with a new shell.

**Independent Test**: Sandbox shell with `alias gp='git push'` before the block; `ali add gp 'echo pull'`, `gp` prints pull; `ali rm gp`; `alias gp` shows `git push` in the same shell.

**Acceptance Scenarios**:

1. **Given** ali's `gp` overrides a shell alias, **When** I `ali rm gp`, **Then** the original alias is restored in the current shell.
2. **Given** ali's `x` did not override anything, **When** I `ali rm x`, **Then** `x` is undefined (current behaviour).
3. **Given** ali's `gp` overrides a shell alias, **When** I `ali update gp ...`, **Then** the new command is active and the original is still remembered for a later removal.

---

### User Story 3 - Check my setup (Priority: P2)

I want one command that tells me whether ali is set up correctly alongside Oh My Zsh and the rest of my `.zshrc`.

**Why this priority**: Order problems (ali block before `oh-my-zsh.sh`, or aliases redefined later in `.zshrc`) are invisible otherwise.

**Independent Test**: `ali doctor` in sandboxes reproducing each problem reports it and exits non-zero; on a healthy setup it reports OK and exits 0.

**Acceptance Scenarios**:

1. **Given** the ali block is missing from `.zshrc`, **When** I run `ali doctor`, **Then** it says so and suggests `ali init zsh`.
2. **Given** `source $ZSH/oh-my-zsh.sh` appears after the ali block, **When** I run `ali doctor`, **Then** it warns that Oh My Zsh will override ali's aliases and says how to fix it.
3. **Given** an ali alias is redefined by something loaded after the block, **When** I run `ali doctor` from the shell, **Then** it lists that alias and its live definition.
4. **Given** the block in `.zshrc` was written by an older ali version, **When** I run `ali doctor`, **Then** it reports the block is outdated and suggests `ali init zsh`.

---

### User Story 4 - Upgrade the shell block (Priority: P2)

When ali's `.zshrc` block changes between versions, `ali init zsh` should replace the old block in place (with a backup) instead of saying "already set up".

**Why this priority**: Stories 1-2 change the block; existing users must be able to get it.

**Independent Test**: A sandbox `.zshrc` containing an older block plus content before and after it; `ali init zsh` replaces only the block, keeps everything else byte-for-byte, makes a backup, and a second run is a no-op.

**Acceptance Scenarios**:

1. **Given** an outdated block, **When** I run `ali init zsh`, **Then** the block is replaced in place, a backup is written, and the output says "updated".
2. **Given** a current block, **When** I run `ali init zsh`, **Then** nothing changes ("already set up").
3. **Given** a block whose end marker is missing (hand-damaged), **When** I run `ali init zsh`, **Then** nothing is changed and I get an error explaining how to fix it.

---

### Edge Cases

- Oh My Zsh updates change a plugin alias after ali recorded it: the remembered original is the one present when the shell loaded, which is what a new shell would also have.
- The same name is both an alias and a function in the shell: report the alias (it takes precedence in zsh).
- A user alias with a name ali reserves (e.g. `command`) is never touched by ali.
- `ali doctor` run outside the wrapper cannot see live aliases: it says which checks were skipped instead of reporting OK.
- `.zshrc` is a symlink: block upgrade writes through it (ADR 0023).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When run through the shell wrapper, `ali add` MUST warn if the name is already an alias or function in the current shell and was not defined by ali, showing its current definition.
- **FR-002**: The override warning MUST NOT block creation and MUST NOT appear for names ali already owns.
- **FR-003**: When run outside the wrapper, `ali` MUST keep the existing builtin/PATH warning and MUST NOT fail.
- **FR-004**: When an ali alias that overrode a pre-existing shell alias is removed or renamed, the live shell MUST restore the pre-existing alias.
- **FR-005**: `ali doctor` MUST check: block present; block after `oh-my-zsh.sh` if Oh My Zsh is used; block current; store valid; and (inside the wrapper) ali aliases not redefined after the block. It MUST exit non-zero if any check fails and MUST say which checks it could not run.
- **FR-006**: `ali init zsh` MUST replace an outdated block in place, keep all other content unchanged, back up first, and refuse to edit a block whose markers are unbalanced.
- **FR-007**: The README MUST document how ali works with Oh My Zsh.

### Key Entities

- **Shell definition report**: what a name currently means in the live shell (alias text, function, or nothing), supplied by the wrapper.
- **Overridden original**: a pre-existing alias that an ali alias replaced in a live shell, remembered for that shell session only.

## Success Criteria *(mandatory)*

- **SC-001**: In tests reproducing the Oh My Zsh git plugin, 100% of overrides of pre-existing aliases or functions produce a warning.
- **SC-002**: After `ali rm` of an override, the original is active in the same terminal with no further action.
- **SC-003**: `ali doctor` correctly flags each of the four misconfigurations in US3, with no false alarms on a healthy setup.
- **SC-004**: Upgrading the block leaves every byte outside the block unchanged.
- **SC-005**: No added shell startup cost beyond 1 ms.

## Assumptions

- Oh My Zsh is sourced from `.zshrc` in the usual `source $ZSH/oh-my-zsh.sh` form; other frameworks (prezto, zinit) get the generic checks only.
- Remembering originals for the session only is enough; ali does not store other tools' aliases in its own store (which may be synced to other machines).
- Shipping ali as an Oh My Zsh plugin is out of scope for this feature.
