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
make test vet fmt-check     # all must pass; CI runs the same on macOS
```

Commits are small, single-purpose and green (`make fmt-check vet test` passes at every
commit), written as [Conventional Commits](https://www.conventionalcommits.org/)
(`feat(cli): ...`, `fix(store): ...`). Hard-to-reverse decisions get an ADR in
[`docs/adr/`](docs/adr) in the same commit. Both are project rules; see the constitution.
