# Feature Specification: Command expansion (sub-aliases, placeholders, environments)

**Feature Branch**: `003-command-expansion`

**Created**: 2026-11-08

**Status**: Draft (awaiting owner approval)

**Input**: User description: "We need argument support. For example, `kubectl get pods -n namespace-1` could turn into `k g p ...`. Aliases have a parent-child relationship: g and p after k belong to k unless specified as global, so g after h is not the same as g after k; a global one can be used under any tool, and the tool's own wins on conflict. Argument lists can be global or tied to a tool, with a good way to view them. Environments: arguments with the same name whose values switch depending on the selected environment." Owner decisions after review: show expansions in the command line by default with a setting for hidden run-time expansion; use named placeholders; environments per terminal by default with a flag to set globally, always visible; use `@` for values typed at call time.

## Clarifications

### Session 2026-11-08

Owner decisions from the design review:

- Q: Should expansion be visible in the command line or happen when the command runs? → A: Visible in the command line by default; a setting switches to run-time expansion.
- Q: Positional argument numbers (`1=1 2=1`) or named placeholders? → A: Named placeholders.
- Q: Is the selected environment per terminal or global? → A: Per terminal by default, with a flag to set it globally; always visible.
- Q: What marks a value typed at call time? → A: `@` (replacing the proposed `-ga`/`-gal` flags, which would collide with the tool's own flags and with `ali add -g`).

Resolved during clarification (recommended options):

- Q: What placeholder syntax avoids clashing with real commands? → A: `{{name}}`, recognised only when `name` is a defined variable. `<name>` is shell redirection; `{name}` collides with awk, `find -exec {}` and `${VAR}`; `{{...}}` collides only with Go templates, and only if a variable of the same name exists.
- Q: In `@var=value`, what if `value` is also a value name? → A: The named value wins; otherwise the text is used literally.
- Q: Do plain aliases without sub-aliases or placeholders now expand in the command line? → A: No. They keep behaving as zsh aliases (FR-017); the command line expands only when a sub-alias, placeholder or `@` word is involved.
- Q: How do I use another environment for just one command? → A: `@@name` in the line (FR-008).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Tool-scoped sub-aliases, expanded as I type (Priority: P1)

I define short words that only mean something after a given tool: under kubectl, `g` is `get` and `p` is `pods`; under helm, `g` can mean something else. When I type `k g p` and a space, my command line turns into `kubectl get pods ` so I can see exactly what I'm about to run. A few words like `o` (`-o wide`) I want everywhere; those are shared, and a tool's own word wins if both exist.

**Why this priority**: It is the core of "argument support" and is useful without variables.

**Independent Test**: With `k` = `kubectl`, kubectl subs `g`=`get`, `p`=`pods`, helm sub `g`=`get values`, shared `o`=`-o wide`: typing `k g p o` then Enter runs and records `kubectl get pods -o wide`; `helm g` becomes `helm get values`; `kubectl g p` (no alias) also expands.

**Acceptance Scenarios**:

1. **Given** the subs above, **When** I type `k g p` followed by a space, **Then** the command line reads `kubectl get pods ` before I press Enter.
2. **Given** a shared sub `g` = `grep` and a kubectl sub `g` = `get`, **When** I type `k g`, **Then** it becomes `kubectl get` (tool wins).
3. **Given** the same subs, **When** I type `k logs mypod p`, **Then** `p` stays `p` (expansion stops at the first word that is not a sub-alias).
4. **Given** the same subs, **When** I type `k g \p`, **Then** the result is `kubectl get p` (backslash keeps a word literal).
5. **Given** I press Enter without typing a space after the last word, **Then** the line is still fully expanded before it runs, and history records the expanded command.
6. **Given** a line that uses no ali alias or sub (e.g. `ls -la`), **When** I type it, **Then** nothing about typing or running it changes.

---

### User Story 2 - Named placeholders and values with @ (Priority: P1)

I put named placeholders into an alias or sub-alias, like `-n {{ns}}`, and keep a list of named values for `ns` (`one` = `namespace-1`, `sys` = `kube-system`). When I run it I pick a value with `@`: `k g p n @sys` runs `kubectl get pods -n kube-system`. I can also type a one-off value, `@ns=my-test-ns`.

**Why this priority**: It is the "arguments" half of the request and the basis for environments.

**Independent Test**: With variable `ns` (values `one`, `sys`; default `one`) and kubectl sub `n` = `-n {{ns}}`: `k g p n @sys` → `kubectl get pods -n kube-system`; `k g p n` → `-n namespace-1` (default); `k g p n @ns=x-y` → `-n x-y`; with no default and no value, the line does not run and I am told what is missing.

**Acceptance Scenarios**:

1. **Given** the setup above, **When** I run `k g p n @sys`, **Then** `kubectl get pods -n kube-system` runs and `@sys` does not appear in the command or history.
2. **Given** two placeholders `{{ns}}` and `{{ctx}}` whose variables both have a value named `dev`, **When** I type `@dev`, **Then** the line does not run and I am asked to write `@ns=dev` or `@ctx=dev`.
3. **Given** `npm i @types/node` or `curl -d @body.json`, **When** I run it, **Then** the `@` words are left untouched (only words that match a placeholder in the line are consumed).
4. **Given** a command containing `{{end}}` (e.g. a Go template) and no variable named `end`, **When** it is saved and run, **Then** `{{end}}` is literal text.
5. **Given** a value containing spaces or shell characters (`a b; rm -rf ~`), **When** it is substituted, **Then** it reaches the command as one literal argument and nothing in it is executed.
6. **Given** I run `ali var ls` or `ali var show ns`, **Then** I see each variable, its values and default, and which aliases/sub-aliases use it.

---

### User Story 3 - Environments (Priority: P2)

I define environments that set values for my variables, e.g. `dev` (`ns` = `namespace-1`, `ctx` = `dev-cluster`) and `prod` (`ns` = `prod`, `ctx` = `prod-cluster`). I switch the environment for this terminal with `ali env use prod`, or for all terminals with `--global`, and my prompt always shows which one is active. Production is marked protected, so a command that uses it waits for a second Enter after showing me the final line.

**Why this priority**: Builds on US2; most valuable for multi-cluster work, but optional for basic use.

**Independent Test**: Two sandbox terminals: `ali env use prod` in one changes `k g p n` there to `-n prod`, the other still uses its own; `ali env use --global dev` changes new terminals and any terminal without a per-terminal choice; the prompt shows the active name; a protected environment needs a second Enter.

**Acceptance Scenarios**:

1. **Given** environments `dev` and `prod`, **When** I run `ali env use prod` in terminal A, **Then** placeholders in terminal A resolve from `prod`, terminal B is unaffected, and A's prompt shows `prod`.
2. **Given** a global environment `dev` and terminal A set to `prod`, **When** I open terminal C, **Then** C uses `dev`; A keeps `prod`.
3. **Given** `ali env use --global staging` run in terminal A, **Then** A switches to `staging` too (its per-terminal choice is cleared) and terminals without their own choice follow.
4. **Given** an explicit `@ns=x` and an active environment that sets `ns`, **When** I run the command, **Then** `@ns=x` wins over the environment, which wins over the variable's default.
5. **Given** the active environment does not define a placeholder's variable and there is no default, **Then** the line does not run and the message names the variable and the environment.
6. **Given** protected `prod` is active, **When** I press Enter on a line that uses a placeholder, **Then** the fully resolved line is shown with a "protected: press Enter again to run" notice and runs only on the second Enter.
7. **Given** `ali env off`, **Then** no environment is active and the prompt indicator disappears.
8. **Given** a one-off `@@prod` in a line, **Then** only that line resolves from `prod`.

---

### User Story 4 - Hidden (run-time) expansion setting (Priority: P2)

I may prefer my command line to stay short. With `ali config set expand runtime`, `k g p n @sys` stays as typed and expands only when it runs.

**Why this priority**: Owner-requested option; the default remains visible expansion.

**Independent Test**: After switching the setting, typing does not change the command line; running `k g p n @sys` still executes `kubectl get pods -n kube-system`; `ali expand 'k g p n @sys'` prints the command without running it.

**Acceptance Scenarios**:

1. **Given** `expand = runtime`, **When** I type `k g p `, **Then** the command line is unchanged and running it executes the expanded command.
2. **Given** `expand = runtime`, **When** a line starts with a tool name rather than an ali alias (`kubectl g p`), **Then** it is not expanded (documented limitation of run-time mode).
3. **Given** any mode, **When** I run `ali expand '<line>'`, **Then** the expanded command is printed and nothing runs.
4. **Given** `expand = runtime` and a protected environment, **When** the command runs, **Then** the resolved command is printed and I must confirm before it runs.

---

### User Story 5 - Discover and complete (Priority: P3)

I can see everything available for a tool in one place and get tab completion for sub-aliases and `@` values.

**Why this priority**: Improves recall ("a good way to view args"); the other stories work without it.

**Independent Test**: `ali show kubectl` lists the tool's aliases, sub-aliases, shared sub-aliases it can use, and the variables they reference; in a sandbox shell, `k g p n @<TAB>` offers `one` and `sys`; `ali <TAB>` completes ali's own commands.

**Acceptance Scenarios**:

1. **Given** the setup from US1-US2, **When** I run `ali show kubectl`, **Then** I see aliases that start kubectl, kubectl sub-aliases with their expansions, shared sub-aliases (marked when overridden by the tool), and variables used.
2. **Given** a placeholder in the line, **When** I press Tab after `@`, **Then** I am offered that variable's value names (and `var=` forms when several placeholders exist).
3. **Given** I press Tab after `k ` or `k g `, **Then** I am offered kubectl's sub-aliases.

---

### Edge Cases

- A sub-alias with the same name as a real subcommand of the tool (e.g. `cp`): the sub-alias wins and `\cp` keeps it literal. ali does not inspect the tool's own subcommands, so it cannot warn about this.
- Sub-alias expansions containing other sub-alias names are not expanded again (no recursion); the expansion is literal text plus placeholders.
- Wrappers before the tool (`sudo k g p`, `time k g p`, `env X=1 k g p`) expand the same way, using the grouping rules of ADR 0008.
- Pasted text is never expanded while pasting; it is expanded on Enter like typed text.
- Quoted words (`'g'`, `"p"`) and words inside quotes are never expanded.
- Placeholders inside quotes in a definition (`echo '{{ns}}'`) are rejected when the alias is saved, because a value cannot be safely inserted there.
- A line using a placeholder but no `@` and no environment/default: not run; the message lists the missing variables. In visible mode the line stays in the editor so I can fix it.
- Multiple uses of the same placeholder in one line resolve to the same value.
- Renaming or removing a variable that is still referenced is refused, with the list of referencing aliases.
- Removing the active environment clears it everywhere with a notice.
- Expansion in non-interactive shells and scripts: none (as with aliases).
- The store from feature 001 (no subs/variables) keeps working unchanged; older ali versions refuse a store that uses the new sections instead of corrupting it.
- Other zsh plugins that bind Space or Enter (e.g. zsh-autosuggestions, zsh-syntax-highlighting, Oh My Zsh): both must keep working.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Users MUST be able to add, list, update and remove sub-aliases scoped to a tool, and shared sub-aliases usable under any tool.
- **FR-002**: The tool for a line MUST be determined from its first word after expanding an ali alias and skipping wrappers, so `k`, `kubectl` and any other alias that runs kubectl share kubectl's sub-aliases.
- **FR-003**: After the tool, consecutive words that are sub-aliases MUST expand (tool's own first, then shared); expansion MUST stop at the first word that is not one; a leading backslash MUST keep a word literal; quoted words MUST never expand.
- **FR-004**: By default, expansion MUST be shown in the command line before the command runs (on Space after a sub-alias, and fully on Enter), and history MUST record the expanded command.
- **FR-005**: A setting MUST switch to run-time expansion, where the typed line is unchanged and expansion happens when the command runs; run-time mode applies to lines that start with an ali alias.
- **FR-006**: Aliases and sub-alias expansions MAY contain named placeholders `{{name}}`; a placeholder MUST only be recognised when `name` is a defined variable; placeholders inside quotes MUST be rejected at save time.
- **FR-007**: Users MUST be able to define variables with named values and an optional default, and to list and inspect them, including where each is used.
- **FR-008**: At call time, `@key` MUST select a value by name when exactly one placeholder in the line has a variable with that key; `@var=value` MUST set a variable (using the named value if `value` is a key, otherwise the literal); `@@env` MUST select an environment for that line only. Ambiguous `@key` MUST stop the line with a message. `@` words that match nothing MUST be left untouched.
- **FR-009**: Value precedence MUST be: explicit `@` value, then line environment (`@@env`), then the terminal's environment, then the global environment, then the variable's default; if none, the line MUST NOT run and the message MUST name the variable.
- **FR-010**: Substituted values MUST be inserted as single literal shell words; they MUST never be interpreted as shell code.
- **FR-011**: Users MUST be able to define, list, inspect, update and remove environments that assign values (named or literal) to variables, and mark an environment as protected.
- **FR-012**: `ali env use <name>` MUST change the environment for the current terminal only; `--global` MUST change it for all terminals that have no per-terminal choice, including the current one; `ali env off` MUST clear it.
- **FR-013**: The active environment MUST be shown in the prompt by default, distinctly when protected; users MUST be able to turn the indicator off or place it in their own theme.
- **FR-014**: Lines that resolve values from a protected environment MUST require a second confirmation (second Enter in visible mode, a prompt in run-time mode) after showing the resolved command.
- **FR-015**: `ali expand '<line>'` MUST print what a line would become, without running it.
- **FR-016**: `ali show <tool>` MUST list the tool's aliases, sub-aliases, applicable shared sub-aliases and referenced variables; tab completion MUST offer sub-aliases after a tool and `@` values when placeholders are present.
- **FR-017**: Lines that use no ali alias, sub-alias or placeholder MUST behave exactly as before.
- **FR-018**: All new definitions MUST be stored in the same human-editable store as aliases, except machine-local settings (expansion mode, global environment, prompt indicator), which MUST NOT be synced with the store.

### Key Entities

- **Tool**: the program a line runs (e.g. `kubectl`), identified by its command name; the scope for sub-aliases.
- **Sub-alias**: a short word that expands to text after a tool; scoped to one tool or shared.
- **Variable**: a named slot (e.g. `ns`) with named values and an optional default, referenced by `{{name}}` placeholders.
- **Environment**: a named set of variable assignments, optionally protected.
- **Active environment**: per terminal, else global, else none.
- **Settings**: machine-local preferences (expansion mode, global environment, prompt indicator).

## Success Criteria *(mandatory)*

- **SC-001**: `kubectl get pods -n kube-system` can be produced by typing at most 12 characters (`k g p n @sys`).
- **SC-002**: In visible mode, 100% of expanded commands appear in full in the command line and in history before or as they run.
- **SC-003**: Typing a line that uses no ali expansion incurs no measurable added delay; a keypress that triggers expansion adds under 15 ms.
- **SC-004**: In tests, 100% of hostile values (quotes, `;`, `$()`, backticks, newlines) reach the command as one literal argument.
- **SC-005**: Switching environment in one terminal never changes the resolved command in another terminal that has its own choice.
- **SC-006**: A user can find every value for a variable, and everything usable after a tool, with one command each.
- **SC-007**: Existing aliases from feature 001 work unchanged after upgrading.

## Assumptions

- zsh's line editor is available (interactive zsh); expansion does not apply to scripts.
- Typing-time expansion coexists with common plugins by chaining rather than replacing their key bindings; any incompatibility found is documented.
- Variable values and environments are not secrets storage; values may reference environment variables (e.g. `$TOKEN`) only by literal text, which the shell then expands as usual inside the user's command, not inside ali.
- Per-directory environments, value lists computed from commands (e.g. live namespace lists), and expansion in other shells are out of scope.
