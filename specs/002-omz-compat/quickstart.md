# Quickstart / Validation: Oh My Zsh compatibility

Sandbox only (`HOME=$(mktemp -d) ZDOTDIR=$HOME`). Simulate Oh My Zsh with a `.zshrc` that defines `alias gp='git push'` and `mk() { make "$@"; }` and a line `source $ZSH/oh-my-zsh.sh` (with `ZSH` pointing to an empty stub file).

1. `ali init zsh`; in an interactive sandbox shell: `ali add gp 'echo pull'` → warning naming `git push`; `gp` prints `pull`.
2. `ali rm gp`; then `alias gp` → `gp='git push'` in the same shell.
3. `ali add mk 'echo mk'` → function warning.
4. Move the ali block above the `oh-my-zsh.sh` line; `ali doctor` → FAIL for order, exit 1.
5. Replace the block with the v1 text; `ali doctor` → outdated; `ali init zsh` → "updated", backup written, rest of file byte-identical.
6. Startup cost: compare `zsh -i -d -c exit` timings before/after (< 1 ms delta).
