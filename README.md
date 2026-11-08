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
| `ali add NAME COMMAND [-g GROUP] [--force]` | Create an alias. Group defaults to the command's first word (`sudo`, `env`, `VAR=x` and paths are skipped). |
| `ali ls [GROUP] [--json]` | List aliases grouped and aligned, or just one group. |
| `ali groups [--json]` | List groups with counts. |
| `ali search TEXT [--json]` | Case-insensitive match on name, command or group. |
| `ali update NAME COMMAND` | Change what an alias runs. |
| `ali regroup NAME GROUP` | Move an alias to another group (sticks across updates). |
| `ali rm NAME...` | Delete aliases. |
| `ali init zsh [--dry-run]` | One-time shell setup; also upgrades the block after an ali update. |
| `ali import [--file PATH] [--dry-run]` | Import existing aliases. |
| `ali doctor` | Check the setup (block present and current, order with Oh My Zsh, store valid, nothing redefining ali's aliases). |
| `ali sub add (TOOL \| --shared) NAME EXPANSION` | Short words that expand after a tool (`update`, `rm`, `ls` too). |
| `ali var add NAME KEY=VALUE... [--default KEY]` | Named values for `{{NAME}}` placeholders (`set`, `unset`, `default`, `rm`, `ls`, `show`). |
| `ali env add NAME VAR=VALUE... [--protect]` | Sets of values; `ali env use NAME [--global]`, `off`, `current`, `ls`, `show`. |
| `ali show TOOL` | Everything usable with a tool: aliases, sub-aliases, variables. |
| `ali expand 'LINE'` | Print what a line expands to, without running it. |
| `ali config get\|set\|ls` | This machine's settings: `expand` (inline/runtime), `prompt` (right/off). |

Errors go to stderr with a non-zero exit code; `--json` makes listings scriptable.

Every command has `--help` with examples, `ali help files` explains what ali stores
where, and man pages are installed with Homebrew (`man ali`, `man ali-add`, `man 7 ali-files`).
From a source checkout, `make man` writes them to `man/`.

## Sub-aliases, placeholders and environments

Make whole commands short without writing shell functions.

**Sub-aliases** are words that only mean something after a tool:

```sh
ali add k kubectl
ali sub add kubectl g get
ali sub add kubectl p pods
ali sub add helm g 'get values'      # helm has its own g
ali sub add --shared o '-o wide'     # works after any tool ali knows
```

Type `k g p o` and press Space or Enter: your command line becomes
`kubectl get pods -o wide` before it runs, and history keeps the full command.
Sub-aliases expand anywhere after the tool (`k logs mypod p`); write `\p` to keep a
word literal, and nothing after a bare `--` expands. See `ali help expansion`.

**Placeholders** take values you pick with `@`:

```sh
ali var add ns one=namespace-1 sys=kube-system --default one
ali sub add kubectl n '-n {{ns}}'

k g p n @sys          # kubectl get pods -n kube-system
k g p n               # kubectl get pods -n namespace-1   (default)
k g p n @ns=my-test   # literal value
```

Values are always inserted as one quoted word, never run as shell code. Tab after
`@` lists the values. See `ali help placeholders`.

**Environments** switch many values at once:

```sh
ali env add dev  ns=one
ali env add prod ns=production --protect
ali env use prod            # this terminal only; --global for every terminal
k g p n @@dev               # one command in another environment
```

The prompt shows `[ali:prod]` (red when protected), and a command that takes values
from a protected environment is shown in full and needs a second Enter. See
`ali help environments`.

Prefer the line to stay as you typed it? `ali config set expand runtime` expands
when the command runs instead (only lines that start with an ali alias).

## Using ali with Oh My Zsh

ali works alongside Oh My Zsh. `ali init zsh` appends its block after
`source $ZSH/oh-my-zsh.sh`, so Oh My Zsh loads first and ali's aliases win where
both define a name.

- **You are told when you override something.** Oh My Zsh's git plugin defines
  `gp='git push'`. `ali add gp 'git pull'` prints
  `warning: gp is already an alias in this shell (git push); ali's alias will take precedence`
  and still creates it. (The `ali` shell function passes your shell's definition to
  ali, because the ali program itself cannot see your shell's aliases.)
- **Removing your version brings theirs back**, immediately: after `ali rm gp`,
  `gp` is `git push` again in the same terminal.
- **`ali doctor`** tells you if the block ended up above `oh-my-zsh.sh`, if something
  later in `.zshrc` redefines one of ali's aliases, or if the block is out of date.

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

## Platform support

| Platform | Status |
|---|---|
| macOS + zsh (Apple Silicon and Intel) | **Supported.** The only tested combination; CI runs on macOS. |
| Linux + zsh | **Untested.** The code compiles and vets clean for Linux, but CI does not run there yet. It should work; please report issues. |
| bash | **Not supported.** `ali init bash` is rejected. The generated file does load in bash, but removals do not take effect for more than the first stale alias (zsh and bash expand array variables differently), and the `.zshrc` block is zsh-only. |
| fish, nushell, PowerShell | **Not supported.** Different alias syntax and rc files. |
| Windows | **Does not build** (file locking uses Unix `flock`). |

Other current limits:

- **No user-written shell functions.** Arguments in the middle of a command are covered by placeholders, but `ali` does not manage arbitrary shell functions.
- **Plugin compatibility.** Expansion is tested with Oh My Zsh (its `magic-space` binding is kept). zsh-autosuggestions and zsh-syntax-highlighting have not been verified yet; `ali` wraps widgets the same way they do, so they are expected to work.
- **Run-time mode** only expands lines that start with an ali alias, and a literal word must be written as `'\p'` (quoted).
- Expansion needs a running `ali` (about 5 ms) only on the keystrokes that can expand; ordinary typing starts no process.
- **zsh startup files.** The block is added to `.zshrc` (or `$ZDOTDIR/.zshrc`), so aliases exist in interactive shells only, not in scripts or non-interactive `zsh -c`.
- **Shadow warnings** cover builtins, `PATH` commands and, when you run `ali add` from your shell, aliases and functions your shell already has (Oh My Zsh plugins included). Scripts that call the `ali` program directly only get the builtin/`PATH` check.
- No tab completion, man page or undo yet, and no sync between machines (the store is a plain file, so a dotfiles repo works; symlinks are respected).
- Reserved names: `ali`, `alias`, `unalias`, `builtin`, `command`, `source`, `local`, `return`, `unset`, `for`, `do`, `done`, `if`, `then`, `elif`, `else`, `fi` cannot be used as alias names (ali's shell code relies on them).

## Notes

- A warning is printed if a new alias shadows an existing command (`ali add ls ...`), but it is still created.
- Environment: `ALI_CONFIG_DIR` overrides the config directory; `ZDOTDIR` is honoured for `.zshrc`.
- Your stored commands may contain secrets, so the config directory is created `0700` and its files `0600`.
- Scope for now: aliases only, zsh only. Functions and other shells are not goals of v1.

## Development

```sh
make test vet fmt-check
```

This project was built with [Spec Kit](https://github.com/github/spec-kit); the spec, plan
and task list are in [`specs/001-alias-manager/`](specs/001-alias-manager) and the project
principles in [`.specify/memory/constitution.md`](.specify/memory/constitution.md).
Design decisions are recorded as ADRs in [`docs/adr/`](docs/adr).
See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
