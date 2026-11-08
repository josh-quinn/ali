# Feature Specification: Alias Manager (`ali`)

**Feature Branch**: `001-alias-manager`

**Created**: 2026-11-08

**Status**: Draft

**Input**: User description: "A CLI tool that quickly creates, lists, updates and deletes shell aliases on a Mac. Aliases persist for re-use. Aliases are organized automatically by tool/group (e.g. helm), searchable, and listable by group. No more editing .zshrc by hand or hunting through an unorganized text file. Intended for public release on GitHub."

## Clarifications

### Session 2026-11-08

- Q: What happens when `ali add` is given a name that already exists? → A: Fail with a message pointing to `ali update`; `--force` overwrites.
- Q: Does `ali import` remove the original alias lines from the shell config? → A: No. It writes a timestamped backup, imports, and prints the original lines so the user can delete them; a later `--remove-originals` is out of scope for v1.
- Q: When `ali update` changes the command, is the group re-inferred? → A: Only if the group was never set explicitly; an explicit group (`-g` or `regroup`) is sticky.
- Q: How does the current shell refresh after a change? → A: A shell function installed by setup wraps `ali` and re-loads the generated aliases after any command that changed data. The hook does not run `ali` at shell startup.
- Q: Should `ali rm` ask for confirmation? → A: No. It is non-interactive; removed aliases can be restored from the store's version history if the user keeps it in git, and `ali rm` prints what it removed.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create an alias on the fly (Priority: P1)

While working, I realise a command I typed (`gp`) doesn't exist. In one short command I define it, and it works immediately in my current terminal and in every future terminal, with no manual edit of any config file.

**Why this priority**: This is the core pain point. Without it nothing else matters.

**Independent Test**: Run `ali add gp 'git pull'`, then run `gp` in the same shell and in a freshly opened shell; both execute `git pull`.

**Acceptance Scenarios**:

1. **Given** `gp` is not defined, **When** I run `ali add gp 'git pull'`, **Then** the alias is saved, assigned to group `git`, and `gp` works in the current shell immediately.
2. **Given** I added an alias earlier, **When** I open a new terminal, **Then** the alias is available without any action from me.
3. **Given** `gp` already exists, **When** I run `ali add gp 'git push'` without `--force`, **Then** nothing changes and I am told the alias exists and how to update it.
4. **Given** I run `ali add hgp 'helm get pods' -g k8s`, **When** the alias is saved, **Then** it is placed in group `k8s` instead of the inferred `helm`.

---

### User Story 2 - See and search what I have (Priority: P1)

I can't remember every alias. I list them organised by group, view just one group, or search by any fragment of name, command or group.

**Why this priority**: Remembering aliases is the second half of the stated problem.

**Independent Test**: With aliases in several groups, `ali ls` shows them grouped and aligned, `ali ls helm` shows only helm, `ali search pods` returns every alias whose name, command or group contains "pods".

**Acceptance Scenarios**:

1. **Given** aliases in groups `git` and `helm`, **When** I run `ali ls`, **Then** I see aliases under group headings, sorted, with names and commands aligned.
2. **Given** the same data, **When** I run `ali ls helm`, **Then** only helm aliases are shown.
3. **Given** the same data, **When** I run `ali groups`, **Then** I see each group with its alias count.
4. **Given** an alias `hgp` = `helm get pods`, **When** I run `ali search pods`, **Then** `hgp` is returned with its group; a search with no match says so and exits non-zero.
5. **Given** any listing command, **When** I add `--json`, **Then** output is machine-readable.

---

### User Story 3 - Update and delete (Priority: P2)

I can change what an alias runs, move it to another group, or remove it, and the change applies at once.

**Why this priority**: Needed to keep the collection tidy, but less urgent than create/see.

**Independent Test**: `ali update gp 'git pull --rebase'` changes behaviour in the current shell; `ali rm gp` makes `gp` undefined again.

**Acceptance Scenarios**:

1. **Given** alias `gp` exists, **When** I run `ali update gp 'git pull --rebase'`, **Then** the command is replaced and the group is re-inferred only if I did not previously set it explicitly.
2. **Given** alias `gp` exists, **When** I run `ali regroup gp tools`, **Then** it moves to group `tools`.
3. **Given** alias `gp` exists, **When** I run `ali rm gp`, **Then** it is deleted and no longer defined in the current or future shells.
4. **Given** alias `nope` does not exist, **When** I run `ali update nope ...` or `ali rm nope`, **Then** I get a clear error and a non-zero exit.

---

### User Story 4 - One-time setup and import of existing aliases (Priority: P2)

I install the tool, run one setup command that adds a single line to my shell config (with a backup), and optionally import the aliases already scattered in my `.zshrc`.

**Why this priority**: Required for first use and for migrating existing users, but done once.

**Independent Test**: In a sandbox home directory, `ali init zsh` is idempotent (second run changes nothing); `ali import` copies alias lines into the store, grouped, leaving the original file intact apart from a timestamped backup.

**Acceptance Scenarios**:

1. **Given** a fresh setup, **When** I run `ali init zsh`, **Then** I am shown exactly what will be added to my shell config, it is added once with a backup, and running it again makes no change.
2. **Given** `.zshrc` contains `alias k=kubectl` and `alias gs='git status'`, **When** I run `ali import`, **Then** both are stored with inferred groups, a backup is created, and the original lines are reported but not removed.
3. **Given** an imported name already exists in the store, **When** importing, **Then** it is skipped and listed as a conflict.

---

### Edge Cases

- Commands containing single quotes, double quotes, `$`, backticks, `&&`, pipes, newlines or unicode are stored and replayed exactly as typed.
- Alias names with spaces, `=`, `/`, leading `-`, or empty names are rejected with a clear message.
- An alias name that shadows an existing system command (e.g. `ls`) prints a warning to stderr; the alias is still created (no prompt).
- Commands that begin with `sudo`, `env VAR=x`, or a path (`./run.sh`, `/usr/bin/foo`) infer the group from the real tool (`sudo helm ...` is `helm`; paths use the file's base name).
- The store file is missing (first run), empty, or hand-edited into invalid syntax: first run works silently; invalid syntax yields an error naming the file and line, and nothing is overwritten.
- Two `ali` processes write at the same time: no data is lost or corrupted.
- The config directory is read-only or the disk is full: the failure is reported and the previous store is unchanged.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to create an alias with a single command taking a name and a command string.
- **FR-002**: The system MUST assign each new alias to a group automatically, defaulting to the first word of the command (after skipping `sudo` and leading `VAR=value` assignments, and using the base name of a path), and MUST allow an explicit override.
- **FR-003**: Newly created, updated, or removed aliases MUST take effect in the current shell session without the user re-sourcing a file manually, once one-time setup is done.
- **FR-004**: Aliases MUST persist across sessions and reboots.
- **FR-005**: Users MUST be able to list all aliases grouped, list a single group, and list groups with counts.
- **FR-006**: Users MUST be able to search aliases by case-insensitive substring across name, command and group.
- **FR-007**: Users MUST be able to update an alias's command, change its group, and delete it.
- **FR-008**: Creating an alias whose name already exists MUST fail unless the user explicitly forces it.
- **FR-009**: The system MUST validate alias names and refuse invalid ones with an explanatory message.
- **FR-010**: The system MUST warn when a new alias name shadows an existing command on the user's system.
- **FR-011**: Listing and search commands MUST support machine-readable output.
- **FR-012**: The system MUST NOT modify the user's shell configuration except during explicit setup or import commands, and MUST back up any file before changing it.
- **FR-013**: Setup MUST be idempotent.
- **FR-014**: The stored data MUST be human-readable and editable by hand, and the tool MUST pick up hand edits.
- **FR-015**: Writes MUST be atomic and safe against concurrent invocations; failure MUST leave the previous data intact.
- **FR-016**: Import MUST read existing alias definitions from the user's shell config, skip conflicts, and not delete the originals.
- **FR-017**: All commands MUST exit with a non-zero status and a message on the error stream on failure.
- **FR-018**: The tool MUST provide built-in help for every command and a version flag.

### Key Entities

- **Alias**: a short name, the command it expands to, and the group it belongs to.
- **Group**: a named collection of aliases, usually the tool they invoke (e.g. `git`, `helm`); derived by default, overridable per alias.
- **Store**: the persisted, human-editable record of all aliases and groups.
- **Shell Hook**: the single line the user adds once so each new shell loads the aliases and the current shell can refresh after a change.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can go from "command not found" to a working, persisted alias in under 5 seconds and one typed command.
- **SC-002**: After any add, update or remove, the change is visible in the current shell with no further user action (once set up).
- **SC-003**: Listing or searching 1,000 aliases feels instant (under 50 ms).
- **SC-004**: Opening a new terminal is not perceptibly slower (added shell startup cost under 5 ms).
- **SC-005**: A user never has to open their shell config file to manage aliases after the one-time setup.
- **SC-006**: 100% of tested special-character commands round-trip exactly when the alias is executed.
- **SC-007**: A new user can install, set up, and create their first alias by following the README in under 3 minutes.

## Assumptions

- Target platform for v1 is macOS with zsh as the interactive shell; other shells and Linux are out of scope but should not be precluded.
- Only aliases are managed in v1; shell functions and environment variables are out of scope.
- Users are comfortable running a command-line tool and adding one line to their shell config via the setup command.
- The tool is distributed as a single prebuilt program (e.g. via Homebrew) and is released publicly under the MIT licence.
- Aliases are personal to one user account; syncing between machines is left to the user's own tools (the store is a plain file suitable for dotfile repos).
