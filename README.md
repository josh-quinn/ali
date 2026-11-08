# ali

Create, find and organise shell aliases from the terminal, without ever opening `.zshrc`.

```console
$ ali add gp 'git pull'
added gp (git): git pull
$ gp            # works immediately, in this shell and every future one
$ ali add hgp 'helm get pods'
$ ali ls
git
  gp   git pull

helm
  hgp  helm get pods
$ ali search pods
helm
  hgp  helm get pods
```

## Why

You type a command that doesn't exist, the terminal says `command not found`, so you
open `.zshrc`, paste an alias, save, re-source, and lose your train of thought. Months
later `.zshrc` is a wall of unrelated aliases you can't remember.

`ali` makes the whole loop one short command, saves it for good, and keeps aliases
**grouped by tool** (the first word of the command) so you can list, search and browse them.

## Install

macOS with zsh. Requires Go 1.24+ to build from source:

```sh
go install github.com/josh-quinn/ali/cmd/ali@latest
```

Homebrew (`brew install josh-quinn/tap/ali`) will be available with the first tagged release.

## Set up (once)

```sh
ali init zsh        # adds a small, clearly marked block to ~/.zshrc (backup first)
```

`ali init zsh --dry-run` shows exactly what would be added. Open a new terminal (or
`source ~/.zshrc`) and you're done. Already have aliases in `.zshrc`? Pull them in:

```sh
ali import --dry-run   # preview
ali import             # copies simple `alias x=...` lines; your .zshrc is not modified
```

`import` is conservative: values it can't read with certainty (multi-line, `$`/backtick
expansion inside double quotes, `alias -g`) are reported, never guessed.

## Commands

| Command | What it does |
|---|---|
| `ali add <name> <command> [-g group] [--force]` | Create an alias. Group defaults to the command's first word (`sudo`, `env`, `VAR=x` and paths are skipped). |
| `ali ls [group] [--json]` | List aliases grouped and aligned, or just one group. |
| `ali groups [--json]` | List groups with counts. |
| `ali search <text> [--json]` | Case-insensitive match on name, command or group. |
| `ali update <name> <command>` | Change what an alias runs. |
| `ali regroup <name> <group>` | Move an alias to another group (sticks across updates). |
| `ali rm <name>...` | Delete aliases. |
| `ali init zsh [--dry-run]` | One-time shell setup. |
| `ali import [--file path] [--dry-run]` | Import existing aliases. |

Errors go to stderr with a non-zero exit code; `--json` makes listings scriptable.

## How it works

- Your aliases live in `~/.config/ali/aliases.toml` (or `$XDG_CONFIG_HOME/ali/`), one
  table per group. It's plain TOML: edit it, diff it, keep it in your dotfiles.
  ```toml
  version = 1

  [helm]
  hgp = "helm get pods"
  ```
- `ali` renders `aliases.zsh` from that file. The block in `.zshrc` just `source`s it,
  so opening a terminal never launches `ali` (startup cost is a few milliseconds).
- A program can't change its parent shell, so the block also defines a tiny `ali()`
  wrapper that re-sources `aliases.zsh` after each run. That's why changes are live
  immediately. `ali` only unaliases names it defined itself; your own aliases are never touched.
- Writes are atomic and file-locked, and a hand-broken `aliases.toml` is reported with
  its location rather than overwritten.
- The only commands that touch `.zshrc` are `init` and `import`, and both back it up first.

## Notes

- A warning is printed if a new alias shadows an existing command (`ali add ls ...`), but it is still created.
- Environment: `ALI_CONFIG_DIR` overrides the config directory; `ZDOTDIR` is honoured for `.zshrc`.
- Scope for now: aliases only, zsh only. Functions, other shells and Linux are not goals of v1.

## Development

```sh
make test vet fmt-check
```

This project was built with [Spec Kit](https://github.com/github/spec-kit); the spec, plan
and task list are in [`specs/001-alias-manager/`](specs/001-alias-manager) and the project
principles in [`.specify/memory/constitution.md`](.specify/memory/constitution.md).
See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
