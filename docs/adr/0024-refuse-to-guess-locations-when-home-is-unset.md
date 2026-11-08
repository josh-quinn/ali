# 0024. Refuse to guess locations when HOME is unset

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Security audit 2026-11-08

## Context

With `HOME` unset (cron, `env -i`, some containers or sudo configurations) the config path became the relative path `.config/ali` and the rc path became `.zshrc`, so `ali add` wrote a store into the current working directory and `ali init` would have edited a `.zshrc` there. Reproduced by running `env -u HOME ali add x ...` in an empty directory. Writing files to attacker- or accident-chosen working directories is a classic path-handling flaw.

## Decision

We will make `store.DefaultDir()` and `shellconf.Path()` return errors instead of relative paths. Resolution order stays `ALI_CONFIG_DIR`, then an absolute `XDG_CONFIG_HOME`, then `$HOME/.config`; `ZDOTDIR` is honoured only if absolute. A relative `XDG_CONFIG_HOME` is ignored as the XDG Base Directory spec requires. `ALI_CONFIG_DIR` is taken as given because it is an explicit instruction.

## Consequences

ali fails loudly (exit 1, with a message naming `ALI_CONFIG_DIR`) in environments without a home directory instead of scattering files. Callers handle an error from both functions.

## Alternatives considered

- **Fall back to the system temp dir**: would silently lose aliases.
- **Use the passwd database**: more code and behaviour differences across systems; the explicit variable is simpler.
