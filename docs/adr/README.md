# Architecture Decision Records

Why `ali` is built the way it is. See [0001](0001-record-architecture-decisions.md) for the
rules; copy [template.md](template.md) to add one. Add or update an ADR in the same commit
as the change it records.

| # | Decision | Status |
|---|----------|--------|
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0002](0002-use-spec-kit-for-spec-driven-development.md) | Use Spec Kit for spec-driven development | Accepted |
| [0003](0003-small-green-conventional-commits.md) | Small, green, conventional commits | Accepted |
| [0004](0004-implement-in-go-as-a-single-static-binary.md) | Implement in Go as a single static binary | Accepted |
| [0005](0005-name-the-command-ali.md) | Name the command ali | Accepted |
| [0006](0006-depend-only-on-cobra-and-go-toml.md) | Depend only on cobra and go-toml | Accepted |
| [0007](0007-release-under-mit-via-goreleaser-and-a-homebrew-tap.md) | Release under MIT via GoReleaser and a Homebrew tap | Accepted |
| [0008](0008-group-aliases-by-the-first-real-word-of-the-command.md) | Group aliases by the first real word of the command | Accepted |
| [0009](0009-store-aliases-as-hand-editable-toml.md) | Store aliases as hand-editable TOML | Accepted |
| [0010](0010-generate-aliases-zsh-and-source-it-from-zshrc.md) | Generate aliases.zsh and source it from .zshrc | Accepted |
| [0011](0011-install-a-static-block-with-an-ali-wrapper-instead-of-eval.md) | Install a static block with an ali() wrapper instead of eval | Accepted |
| [0012](0012-track-loaded-alias-names-so-removals-take-effect.md) | Track loaded alias names so removals take effect | Accepted |
| [0013](0013-quote-alias-commands-with-single-quotes.md) | Quote alias commands with single quotes | Accepted |
| [0014](0014-write-atomically-under-an-advisory-lock.md) | Write atomically under an advisory lock | Accepted |
| [0015](0015-never-prompt-fail-on-collisions-unless-force.md) | Never prompt; fail on collisions unless --force | Accepted |
| [0016](0016-warn-but-do-not-block-when-an-alias-shadows-a-command.md) | Warn but do not block when an alias shadows a command | Accepted |
| [0017](0017-make-import-conservative-and-non-destructive.md) | Make import conservative and non-destructive | Accepted |
| [0018](0018-back-up-user-files-before-changing-them.md) | Back up user files before changing them | Accepted |
