# ali

[![ci](https://github.com/josh-quinn/ali/actions/workflows/ci.yml/badge.svg)](https://github.com/josh-quinn/ali/actions/workflows/ci.yml)
[![codeql](https://github.com/josh-quinn/ali/actions/workflows/codeql.yml/badge.svg)](https://github.com/josh-quinn/ali/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/josh-quinn/ali/badge)](https://scorecard.dev/viewer/?uri=github.com/josh-quinn/ali)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/16489/badge)](https://www.bestpractices.dev/projects/16489)

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

And beyond plain aliases, short words that expand after a tool, with values you pick:

```console
$ ali add k kubectl
$ ali var add ns dev=dev-ns sys=kube-system --default dev
$ ali sub add kubectl g get; ali sub add kubectl p pods; ali sub add kubectl n '-n {{ns}}'
$ k g p n @sys        # press Enter: the line becomes "kubectl get pods -n kube-system"
```

## Why

You type a command that doesn't exist, the terminal says `command not found`, so you
open `.zshrc`, paste an alias, save, re-source, and lose your train of thought. Months
later `.zshrc` is a wall of unrelated aliases you can't remember.

`ali` makes the whole loop one short command, saves it for good, and keeps aliases
**grouped by tool** (the first word of the command) so you can list, search and browse them.

## Installation and setup

### 1. Requirements

- **macOS** (Apple Silicon or Intel) with **zsh**, the default shell since macOS
  Catalina. Check with `echo $SHELL`; it should end in `/zsh`. Linux with zsh works
  and is tested in CI, but macOS is the primary platform.
- To build from source: **Go 1.24 or newer** (`go version`).
- Optional: Oh My Zsh, fzf, zsh-autosuggestions and similar plugins; see
  [Using ali with Oh My Zsh](#using-ali-with-oh-my-zsh).

### 2. Install the `ali` program

Pick one:

| Method | Command | Notes |
|---|---|---|
| Homebrew | `brew install --cask josh-quinn/tap/ali` | Installs the man pages too. |
| Release download | [GitHub Releases](https://github.com/josh-quinn/ali/releases) | Archives for macOS and Linux (arm64, amd64) with `checksums.txt`. |
| Go | `go install github.com/josh-quinn/ali/cmd/ali@latest` | Puts `ali` in `$(go env GOPATH)/bin`, usually `~/go/bin`. |
| From source | `git clone https://github.com/josh-quinn/ali && cd ali && make build` | Produces `./ali`; copy it somewhere on your `PATH`, e.g. `cp ali ~/.local/bin/`. `make man` builds the man pages into `man/`. |

Release archives carry build-provenance attestations and signed SLSA level 3
provenance (`ali.intoto.jsonl` on each release), so you can check that a download was
built by this repository's release workflow:

```sh
gh attestation verify ali_<version>_darwin_arm64.tar.gz --repo josh-quinn/ali
# or, with slsa-verifier:
slsa-verifier verify-artifact ali_<version>_darwin_arm64.tar.gz \
  --provenance-path ali.intoto.jsonl --source-uri github.com/josh-quinn/ali --source-tag v<version>
```

Then make sure your shell can find it:

```sh
command -v ali      # should print a path
ali --version
```

If `command -v ali` prints nothing, add the directory to your `PATH` in `~/.zshrc`
(for example `export PATH="$HOME/go/bin:$PATH"`) and open a new terminal.

### 3. Connect it to your shell (once)

```sh
ali init zsh --dry-run   # show exactly what would be added, change nothing
ali init zsh             # add it
```

`ali init zsh`:

- writes a timestamped backup next to your `.zshrc` (`.zshrc.ali-backup-<date>`) first;
- appends one clearly marked block (`# >>> ali >>>` … `# <<< ali <<<`) to the end of
  `~/.zshrc` (or `$ZDOTDIR/.zshrc`), so it loads after Oh My Zsh and other plugins;
- follows a symlinked `.zshrc` (stow, chezmoi, a dotfiles repo) instead of replacing it;
- creates `~/.config/ali/` (or `$XDG_CONFIG_HOME/ali`) with private permissions;
- does nothing if the block is already there and current.

The block only loads a file and defines a small `ali` shell function. It never starts
the `ali` program when a terminal opens, so it adds no noticeable startup time.

Now **open a new terminal** (or run `source ~/.zshrc`).

### 4. Check that it works

```sh
ali doctor
```

All five checks should say `ok`. Then try it:

```sh
ali add hi 'echo hello from ali'
hi                      # works immediately, in this terminal
ali ls
ali rm hi
```

If `hi` says `command not found`, see [Troubleshooting](#troubleshooting).

### 5. Bring in the aliases you already have (optional)

```sh
ali import --dry-run    # preview what would be copied
ali import              # copy simple `alias name=value` lines from ~/.zshrc
ali import --file ~/.oh-my-zsh/custom/aliases.zsh   # or from another file
```

Your files are never changed: `import` prints the original lines so you can delete
them yourself once you're happy. It only copies what it can read with certainty;
multi-line values, `$` or backticks inside double quotes, `alias -g`, and several
definitions on one line are reported, never guessed.

### 6. Tab completion (optional)

`ali init zsh` also installs Tab completion for ali's own commands. It is active when
zsh's completion system is loaded, which Oh My Zsh does for you. Without Oh My Zsh, make
sure your `.zshrc` has `autoload -Uz compinit && compinit` **above** the ali block.
Completion of `@` values and sub-aliases works either way.

### 7. Settings (optional)

```sh
ali config ls                     # expand = inline, prompt = right
ali config set expand runtime     # expand when the command runs, not on screen
ali config set prompt off         # don't touch the right prompt
```

See `ali help expansion` and `ali help environments`.

### Keeping your aliases in a dotfiles repo

Everything you define lives in `~/.config/ali/aliases.toml`, a plain TOML file you can
read, edit, diff and commit. To sync it, keep the file in your dotfiles repo and
symlink it to `~/.config/ali/aliases.toml`; ali writes through the symlink. Do **not**
sync `settings.toml` or `global-env`: they belong to one machine. `ali help files`
lists every file ali uses.

### Upgrading

Install the new version the same way you installed it, then run:

```sh
ali init zsh      # updates the block if this version changed it (with a backup)
ali doctor        # "block up to date" should say ok
```

### Uninstalling

1. Delete the block between `# >>> ali >>>` and `# <<< ali <<<` from `~/.zshrc`, or
   restore the backup that `ali init` made.
2. Remove the program (`brew uninstall --cask ali`, or delete the binary).
3. Optionally delete your aliases: `rm -r ~/.config/ali`.

### Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `zsh: command not found: ali` | The binary is not on your `PATH`; see step 2. |
| A new alias works only in new terminals | The `ali` shell function is not loaded. Run `ali doctor` (or `whence -w ali`, which should say `function`) and `ali init zsh`, then open a new terminal. |
| `ali doctor`: block loads after Oh My Zsh = FAIL | The block is above `source $ZSH/oh-my-zsh.sh`; move it below, so Oh My Zsh's aliases don't replace ali's. |
| `ali doctor`: aliases redefined after the block | Something later in `.zshrc` (or a plugin loaded after it) redefines that alias. Remove it there or keep the ali block last. |
| `ali doctor`: block up to date = FAIL | Run `ali init zsh` after upgrading. |
| `ali: cannot locate your config directory` | `HOME` is not set (cron, `env -i`). Set `ALI_CONFIG_DIR` to choose a directory. |
| Sub-aliases do not expand on Space | Check `ali config get expand` (`inline`), that the line starts with an ali alias or a tool that has sub-aliases (`ali show TOOL`), and that you are in an interactive terminal. |
| A word expanded that should not have | Write it as `\word`, quote it, or put it after a bare `--`. |

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
| `ali help TOPIC` | Guides for what you type in the shell: `expansion`, `placeholders`, `environments`, and `files`. |

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

- Everything you define lives in `~/.config/ali/aliases.toml` (or
  `$XDG_CONFIG_HOME/ali/`): aliases in one table per group, plus `_subs`, `_shared`,
  `_vars` and `_envs` tables once you use those features. It's plain TOML: edit it,
  diff it, keep it in your dotfiles.
  ```toml
  version = 2

  [helm]
  hgp = "helm get pods"

  [_subs.kubectl]
  g = "get"
  n = "-n {{ns}}"

  [_vars.ns]
  default = "dev"
  [_vars.ns.values]
  dev = "dev-ns"
  ```
  The file stays `version = 1` until you use sub-aliases, variables or environments,
  so older versions of ali keep reading it.
- `ali` renders `aliases.zsh` from that file on every change. The block in `.zshrc`
  only `source`s it, so opening a terminal never launches the `ali` program.
- A program can't change its parent shell, so the block also defines a small `ali()`
  shell function that reloads `aliases.zsh` after each command. That's why changes
  are live immediately. ali only unaliases names it defined itself; your own aliases
  are never touched.
- Expansion while you type comes from small zsh line-editor widgets in `aliases.zsh`.
  They check, inside zsh, whether a line can involve an ali expansion and only then
  ask `ali expand` (about 5 ms). Ordinary typing never starts a process. They wrap
  zsh's own widgets rather than rebinding keys, so plugins such as Oh My Zsh and fzf
  keep their bindings.
- An alias containing a placeholder is defined as a small function, so it still
  expands correctly where the line editor is not involved.
- Writes are atomic and file-locked, files are private (`0700`/`0600`), symlinks are
  followed, and a hand-broken `aliases.toml` is reported with its location rather
  than overwritten.
- `ali init` is the only command that changes `.zshrc`, and it backs it up first.
  `ali import` only reads it.

## Platform support

| Platform | Status |
|---|---|
| macOS + zsh (Apple Silicon and Intel) | **Supported.** The primary platform; CI runs every test, including the real-zsh end-to-end tests, on macOS. |
| Linux + zsh | **Tested in CI** (Ubuntu, same test suite). Less used day to day than macOS; please report issues. |
| bash | **Not supported.** `ali init bash` is rejected. The generated file does load in bash, but removals do not take effect for more than the first stale alias (zsh and bash expand array variables differently), and the `.zshrc` block is zsh-only. |
| fish, nushell, PowerShell | **Not supported.** Different alias syntax and rc files. |
| Windows | **Does not build** (file locking uses Unix `flock`). |

Other current limits:

- **No user-written shell functions.** Arguments in the middle of a command are covered by placeholders, but `ali` does not manage arbitrary shell functions.
- **Plugin compatibility.** Tested with Oh My Zsh (its `magic-space` binding is kept) and fzf (its Tab and Ctrl-R/Ctrl-T bindings keep working). zsh-autosuggestions and zsh-syntax-highlighting have not been verified yet; `ali` wraps widgets the same way they do, so they are expected to work.
- **Run-time mode** only expands lines that start with an ali alias, and a literal word must be written as `'\p'` (quoted).
- Expansion needs a running `ali` (about 5 ms) only on the keystrokes that can expand; ordinary typing starts no process.
- **zsh startup files.** The block is added to `.zshrc` (or `$ZDOTDIR/.zshrc`), so aliases exist in interactive shells only, not in scripts or non-interactive `zsh -c`.
- **Shadow warnings** cover builtins, `PATH` commands and, when you run `ali add` from your shell, aliases and functions your shell already has (Oh My Zsh plugins included). Scripts that call the `ali` program directly only get the builtin/`PATH` check.
- No undo, and no built-in sync between machines (the store is a plain file, so a dotfiles repo works; symlinks are respected).
- Reserved names: `ali`, `alias`, `unalias`, `builtin`, `command`, `source`, `local`, `return`, `unset`, `for`, `do`, `done`, `if`, `then`, `elif`, `else`, `fi` cannot be used as alias names (ali's shell code relies on them).

## Notes

- A warning is printed if a new alias shadows an existing command (`ali add ls ...`), but it is still created.
- Environment: `ALI_CONFIG_DIR` overrides the config directory; `ZDOTDIR` is honoured for `.zshrc`.
- Your stored commands may contain secrets, so the config directory is created `0700` and its files `0600`.
- Scope: aliases, sub-aliases, placeholders and environments, on zsh. Arbitrary shell functions and other shells are not goals for now.

## Development

```sh
make build          # ./ali
make check          # everything CI checks: fmt, vet, golangci-lint, tests, govulncheck
make man            # man pages into man/
make hooks          # once per clone: scan every push for secrets (needs gitleaks)
```

Linting needs [golangci-lint](https://golangci-lint.run) v2 and
[ShellCheck](https://www.shellcheck.net) (`brew install golangci-lint shellcheck`),
and the secret scan needs [gitleaks](https://github.com/gitleaks/gitleaks)
(`brew install gitleaks`). With `make hooks`, every `git push` first scans the commits
it would publish and is blocked if anything looks like a token, key or password
([ADR 0041](docs/adr/0041-scan-for-secrets-before-every-push-and-again-in-ci.md));
CI scans the full history again.
The end-to-end tests drive a real zsh through `zsh/zpty` in a throwaway home directory,
so they never touch your own shell configuration.

Every pull request runs, on macOS and Linux with Go 1.24 and the latest Go: the full
test suite with the race detector, gofmt, go vet, golangci-lint (with gosec),
govulncheck, gitleaks (full history), CodeQL, OpenSSF Scorecard, dependency review, actionlint and
`goreleaser check`. All of these are free for open-source projects
([ADR 0040](docs/adr/0040-gate-every-change-with-free-open-source-quality-and-security-checks.md)).

This project was built with [Spec Kit](https://github.com/github/spec-kit); specs,
plans and task lists are in [`specs/`](specs) and the project principles in
[`.specify/memory/constitution.md`](.specify/memory/constitution.md).
Design decisions are recorded as ADRs in [`docs/adr/`](docs/adr).
See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

Please report vulnerabilities privately; see [SECURITY.md](SECURITY.md).

## License

MIT
