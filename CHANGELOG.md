# Changelog

All notable changes are listed here; the GitHub Release for each version also lists
every commit, grouped by type.

## Unreleased

## 0.1.2 - 2026-11-09

- Release automation: GoReleaser only accepts a bare `{{ .Env.NAME }}` as the
  Homebrew token, so 0.1.1 was also published without its cask, attestations or
  SLSA provenance. Fixed, with a test (`tests/release`) for the publish-time fields
  that `goreleaser check` cannot validate. The program is unchanged from 0.1.0.

## 0.1.1 - 2026-11-09

- Release automation: the Homebrew cask step used a template function GoReleaser
  does not have, so v0.1.0 was published without the cask, attestations or SLSA
  provenance. Fixed; 0.1.1 is the first complete release. The program itself is
  unchanged from 0.1.0.

## 0.1.0 - 2026-11-09

First public release.

- Releases: GitHub Releases for macOS and Linux (arm64, amd64) with checksums,
  build-provenance attestations and man pages; Homebrew cask in `josh-quinn/tap`
  (ADRs 0039, 0042).
- Quality and security gates in CI on macOS and Linux, and a pre-push secret scan
  (`make hooks`) (ADRs 0040, 0041).
- Command expansion (feature 003): tool-scoped and shared sub-aliases (`ali sub`), `{{name}}` placeholders with named values (`ali var`) picked with `@key` / `@var=value`, environments (`ali env`, per terminal or `--global`, `@@env` for one line, shown in the prompt, protected ones need a second Enter), visible expansion in the command line by default or run-time expansion (`ali config set expand runtime`), `ali expand`, `ali show`, Tab completion of values, sub-aliases and ali's own commands, and help topics `expansion`, `placeholders`, `environments` (ADRs 0030-0037). Store format v2 is written only when these are used.
- The `.zshrc` block is now v3; run `ali init zsh` to upgrade (`ali doctor` reports an old block).
- Oh My Zsh compatibility (feature 002): `ali add` warns when overriding an alias or function already in your shell; removing ali's version restores the original in the same terminal; new `ali doctor`; `ali init zsh` upgrades an outdated block in place (ADRs 0027-0029).
- Documentation: every command has a description and examples in `--help`, `ali help files`, and generated man pages shipped with releases (ADR 0038).
- `if`, `then`, `elif`, `else`, `fi` are now reserved alias names.
- Security: reserve names the shell integration relies on and escape its builtins (ADR 0021); private config permissions (0022); write through symlinks (0023); fail instead of using relative paths without HOME (0024); escape control characters in output and reject NUL (0025); importer skips unquoted shell metacharacters; CI hardened and Dependabot enabled (0026).
- Initial implementation: `add`, `ls`, `groups`, `search`, `update`, `regroup`, `rm`, `init zsh`, `import`.
