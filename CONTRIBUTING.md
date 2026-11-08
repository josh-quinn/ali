# Contributing

Thanks for helping. A few ground rules, taken from the project [constitution](.specify/memory/constitution.md):

- **Test first.** Add a failing test, then the code. Shell quoting, grouping, atomic writes
  and import parsing need explicit edge-case tests.
- **Never touch the real shell config in tests.** Use a temp `HOME`/`ALI_CONFIG_DIR`
  (see `tests/e2e`).
- **Keep it small and fast.** No new dependencies without justification; reads must stay
  well under 10 ms.
- **Stay in scope.** v1 is aliases on zsh. Larger features should start as an issue and a
  spec under `specs/`.

```sh
brew install golangci-lint gitleaks shellcheck
make hooks     # once per clone: every git push is scanned for secrets first
make check     # fmt, vet, golangci-lint, tests, govulncheck, gitleaks: what CI runs
```

If the pre-push scan blocks you, remove the secret from **history** (amend or rebase;
deleting it in a new commit is not enough, because history is pushed too) and rotate
the credential. Record genuine false positives by fingerprint in `.gitleaksignore`.

CI also runs CodeQL, OpenSSF Scorecard, dependency review, actionlint and
`goreleaser check`, on macOS and Linux with Go 1.24 and the latest Go.

Commits are small, single-purpose and green (`make fmt-check vet test` passes at every
commit), written as [Conventional Commits](https://www.conventionalcommits.org/)
(`feat(cli): ...`, `fix(store): ...`). Hard-to-reverse decisions get an ADR in
[`docs/adr/`](docs/adr) in the same commit. Both are project rules; see the constitution.
