# CLI Contract

Global: `ali --help`, `ali --version`, every command has `-h`. Errors → stderr, exit 1 (usage errors exit 2). Write commands print a one-line confirmation to stdout.

| Command | Behaviour | Output / exit |
|---------|-----------|---------------|
| `ali add <name> <command> [-g group] [--force]` | create; warn on shadowing | `added gp (git): git pull`; exit 1 if exists and no `--force` |
| `ali ls [group] [--json]` | list grouped, sorted, aligned | groups with headings; `--json` = `[{"name","command","group"}]`; exit 1 if group unknown |
| `ali groups [--json]` | groups + counts | `git  3` per line |
| `ali search <text> [--json]` | case-insensitive substring over name, command, group | same shape as `ls`; exit 1 when no matches |
| `ali update <name> <command>` | replace command | `updated gp`; exit 1 if missing |
| `ali regroup <name> <group>` | move and pin | `moved gp to tools` |
| `ali rm <name>...` | delete | `removed gp: git pull`; exit 1 if any missing (others still removed) |
| `ali init zsh [--dry-run]` | add marked block to `.zshrc` with backup; idempotent | shows block; `already set up` on repeat |
| `ali import [--file path] [--dry-run]` | import `alias` lines, backup, skip conflicts | summary `imported N, skipped M`; lists original lines to delete |

Files: store `~/.config/ali/aliases.toml`, generated `aliases.zsh`, lock `aliases.toml.lock`, backups `.zshrc.ali-backup-<timestamp>`. Environment: `XDG_CONFIG_HOME`, `ALI_CONFIG_DIR` (override, used by tests), `ZDOTDIR` honoured for `.zshrc` location.
