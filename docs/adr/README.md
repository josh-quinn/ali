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
| [0007](0007-release-under-mit-via-goreleaser-and-a-homebrew-tap.md) | Release under MIT via GoReleaser and a Homebrew tap | Superseded by [0039](0039-distribute-with-a-homebrew-cask-built-by-goreleaser.md) |
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
| [0019](0019-limit-v1-scope-to-aliases-on-zsh-macos-first.md) | Limit v1 scope to aliases on zsh (macOS first) | Accepted |
| [0020](0020-test-end-to-end-through-a-line-fed-interactive-zsh.md) | Test end to end through a line-fed interactive zsh | Accepted |
| [0021](0021-reserve-names-the-shell-integration-relies-on-and-escape-its-builtins.md) | Reserve names the shell integration relies on and escape its builtins | Accepted |
| [0022](0022-keep-the-config-directory-and-its-files-private.md) | Keep the config directory and its files private | Accepted |
| [0023](0023-write-through-symlinks-instead-of-replacing-them.md) | Write through symlinks instead of replacing them | Accepted |
| [0024](0024-refuse-to-guess-locations-when-home-is-unset.md) | Refuse to guess locations when HOME is unset | Accepted |
| [0025](0025-escape-control-characters-when-printing-stored-text-and-reject-nul.md) | Escape control characters when printing stored text and reject NUL | Accepted |
| [0026](0026-harden-ci-and-keep-dependencies-patched.md) | Harden CI and keep dependencies patched | Superseded by [0040](0040-gate-every-change-with-free-open-source-quality-and-security-checks.md) |
| [0027](0027-learn-about-the-live-shell-through-the-wrapper.md) | Learn about the live shell through the wrapper | Accepted |
| [0028](0028-remember-overridden-aliases-in-the-shell-session.md) | Remember overridden aliases in the shell session | Accepted |
| [0029](0029-upgrade-the-managed-block-in-place.md) | Upgrade the managed block in place | Accepted |
| [0030](0030-show-expansions-in-the-command-line-by-default.md) | Show expansions in the command line by default | Accepted |
| [0031](0031-scope-sub-aliases-to-the-tool-and-expand-them-anywhere-in-the-command.md) | Scope sub-aliases to the tool and expand them anywhere in the command | Accepted |
| [0032](0032-use-named-name-placeholders-recognised-only-for-defined-variables.md) | Use named {{name}} placeholders recognised only for defined variables | Accepted |
| [0033](0033-use-for-call-time-values-and-for-a-one-off-environment.md) | Use @ for call-time values and @@ for a one-off environment | Accepted |
| [0034](0034-choose-environments-per-terminal-by-default-globally-by-flag-and-always-show-them.md) | Choose environments per terminal by default, globally by flag, and always show them | Accepted |
| [0035](0035-one-expansion-engine-in-go-behind-thin-gated-zsh-widgets.md) | One expansion engine in Go behind thin, gated zsh widgets | Accepted |
| [0036](0036-insert-substituted-values-as-single-literal-words.md) | Insert substituted values as single literal words | Accepted |
| [0037](0037-store-format-v2-only-when-used-with-machine-local-settings-kept-separate.md) | Store format v2 only when used, with machine-local settings kept separate | Accepted |
| [0038](0038-generate-help-and-man-pages-from-the-command-tree.md) | Generate help and man pages from the command tree | Accepted |
| [0039](0039-distribute-with-a-homebrew-cask-built-by-goreleaser.md) | Distribute with a Homebrew cask built by GoReleaser | Accepted |
| [0040](0040-gate-every-change-with-free-open-source-quality-and-security-checks.md) | Gate every change with free, open-source quality and security checks | Accepted |
| [0041](0041-scan-for-secrets-before-every-push-and-again-in-ci.md) | Scan for secrets before every push and again in CI | Accepted |
| [0042](0042-release-from-version-tags-with-goreleaser-and-build-provenance.md) | Release from version tags with GoReleaser and build provenance | Accepted |
