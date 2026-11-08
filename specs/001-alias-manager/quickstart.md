# Quickstart / Validation Guide

Prerequisites: Go toolchain, zsh. Run in a sandbox so the real `~/.zshrc` is untouched.

```sh
go build -o /tmp/ali-bin/ali ./cmd/ali
export PATH=/tmp/ali-bin:$PATH
export HOME=$(mktemp -d) ZDOTDIR=$HOME
printf "alias k=kubectl\n" > $HOME/.zshrc

ali init zsh          # block added once; backup created; rerun → "already set up"
ali import            # imports k under group kubectl
# interactive shells read stdin line by line (zsh -c would parse the whole script up front):
printf 'ali add gp "echo pulled"\ngp\nexit\n' | zsh -i   # prints "pulled" in the same shell
ali add hgp 'helm get pods'; ali ls; ali ls helm; ali groups; ali search pods
ali update gp 'git pull --rebase'; ali regroup gp tools; ali rm gp
ali add q "echo 'it'\''s \$HOME'"; printf 'q\nexit\n' | zsh -i   # quoting round-trip
```
Expected: matches acceptance scenarios in spec.md; `go test ./...` passes; `.zshrc` differs from the original only by the marked block.
