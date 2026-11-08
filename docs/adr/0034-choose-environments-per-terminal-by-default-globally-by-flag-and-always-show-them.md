# 0034. Choose environments per terminal by default, globally by flag, and always show them

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: owner decision 3

## Context

Environments switch variable values (dev vs prod). A global switch silently changes every open terminal, the classic way to run a command against the wrong cluster.

## Decision

We will make `ali env use <name>` apply to the current terminal (the wrapper exports `ALI_ENV`), `--global` set a machine-local default that terminals without their own choice follow (and clear the current terminal's override), and `ali env off` clear it. Precedence: `@` value, `@@env`, terminal, global, default. The active environment is shown in the prompt by a hook that runs no process, in red when protected; protected environments require a second Enter (or a confirmation in run-time mode) after showing the resolved line.

## Consequences

Mistakes are local and visible. The global setting lives outside the synced store, so it does not follow the user to other machines. Users with themes that use `RPROMPT` may prefer `prompt = off` plus `ali_prompt_info`.

## Alternatives considered

- **Global only**: unsafe.
- **Per directory (direnv style)**: useful but out of scope for now.
