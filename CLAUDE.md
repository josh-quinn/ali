# ali: instructions for AI coding agents

Read `.specify/memory/constitution.md` first; it is the source of truth. The rules you
must follow on every change:

1. **Small, green, conventional commits.** One purpose per commit, `type(scope): subject`,
   tests committed with the code, `make fmt-check vet test` passes at every commit.
2. **ADRs.** Record any hard-to-reverse or interface/security/dependency/scope decision in
   `docs/adr/` in the same commit, and keep `docs/adr/README.md` current. Supersede, never
   rewrite, accepted ADRs.
3. **Test first**, including a regression test for every bug and security fix.
4. **Never touch the real shell config in tests.** Use a temp `HOME` / `ALI_CONFIG_DIR`.
5. **No secrets in history.** Run `make hooks` once per clone; never bypass the
   pre-push scan with `--no-verify`.
6. **Scope**: aliases only, zsh on macOS, minimal dependencies (see ADRs).

Workflow: spec-kit (`specs/001-alias-manager/`). Commands: `make build test vet fmt-check`.
