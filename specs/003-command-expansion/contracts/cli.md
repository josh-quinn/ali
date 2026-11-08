# CLI Contract: Command expansion

Errors to stderr, exit 1 (usage 2), `--json` on every listing, as in feature 001.

## Sub-aliases
| Command | Behaviour |
|---------|-----------|
| `ali sub add <tool> <name> <expansion> [--force]` | add a tool sub-alias; `--shared` instead of `<tool>` adds a shared one |
| `ali sub update <tool>\|--shared <name> <expansion>` | replace an expansion |
| `ali sub rm <tool>\|--shared <name>...` | remove |
| `ali sub ls [<tool>\|--shared] [--json]` | list, grouped by tool |

## Variables
| Command | Behaviour |
|---------|-----------|
| `ali var add <name> [key=value...] [--default key]` | create |
| `ali var set <name> key=value...` / `ali var unset <name> key...` | add/change or remove values |
| `ali var default <name> <key>\|--none` | set or clear default |
| `ali var rm <name>` | refused while referenced (lists references) |
| `ali var ls [--json]` / `ali var show <name>` | list; show values, default, environments that set it, aliases/sub-aliases that use it |

## Environments
| Command | Behaviour |
|---------|-----------|
| `ali env add <name> [var=value...] [--protect]` | create |
| `ali env set <name> var=value...` / `ali env unset <name> var...` | change assignments |
| `ali env protect <name>` / `ali env unprotect <name>` | toggle protection |
| `ali env rm <name>` | remove; clears the global choice if it was that environment; terminals that chose it fall back to none with a warning on next expansion |
| `ali env ls [--json]` / `ali env show [<name>]` | list (active marked); show assignments resolved to values |
| `ali env use <name> [--global]` | per terminal (via wrapper) or global |
| `ali env off [--global]` | clear |
| `ali env current` | print active environment and where it came from (`terminal`, `global`, `none`) |

## Settings, expansion, discovery
| Command | Behaviour |
|---------|-----------|
| `ali config get <key>` / `ali config set <key> <value>` | `expand` = `inline`\|`runtime`; `prompt` = `right`\|`off`; changing `expand` regenerates `aliases.zsh` |
| `ali expand '<line>'` | print the expanded command; exit 1 with a message if unresolved or ambiguous; never runs anything |
| `ali expand --run <words...>` | internal (run-time functions): print ali-quoted command for `eval`; for a protected environment, confirm on `/dev/tty`, or exit 1 if there is no terminal |
| `ali expand --zle --cursor N -- <buffer>` | internal (widgets): NUL-separated `status, cursor, buffer, message` (research R4) |
| `ali show <tool> [--json]` | aliases, sub-aliases, applicable shared sub-aliases, referenced variables |

## Call-time syntax (typed in the shell)
| Syntax | Meaning |
|--------|---------|
| `\word` | keep `word` literal (no sub-alias expansion) |
| `{{name}}` | placeholder (in definitions) |
| `@key` | pick value `key` for the only placeholder whose variable has it |
| `@var=value` | set `var` to value key or literal |
| `@@env` | use environment `env` for this line |
