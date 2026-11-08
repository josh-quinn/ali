# 0031. Scope sub-aliases to the tool and expand them anywhere in the command

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: specs/003-command-expansion FR-001-FR-003; owner review

## Context

The owner wants `g` after `k` (kubectl) to differ from `g` after `h` (helm), with shared words usable anywhere and tool words winning. A first draft stopped expanding at the first word that was not a sub-alias, so `k get p` would leave `p` literal because `get` is not a sub-alias. The owner rejected that: expansion should not stop, and users can keep a word literal with a backslash when they need to.

## Decision

We will key sub-aliases by the tool the line runs (first word after alias expansion and wrappers, as in ADR 0008), so every alias for kubectl shares kubectl's sub-aliases. Every plain, unquoted word after the tool that names a sub-alias expands (tool's own first, then shared), wherever it appears in that command, including after literal words such as `get` or a pod name. `\word` and quoted words stay literal; expansions are not re-expanded; each command after `;`, `|`, `&&`, `||` or `&` is scoped to its own tool. A bare `--` ends sub-alias expansion for the rest of that command, because what follows belongs to another program (`kubectl exec pod -- g`).

## Consequences

Sub-aliases work in any position (`k get p`, `k logs mypod p`). A real argument that happens to equal a sub-alias name is expanded too; visible expansion (ADR 0030) shows this before the command runs, and `\p` keeps it literal.

## Alternatives considered

- **Stop at the first non-sub-alias word**: rejected by the owner (`k get p` would not expand).
- **Key by alias name (`k`)**: `kubectl g p` and other kubectl aliases would not share words.
- **Nested tree (verb then noun)**: more to define; flat per-tool scope matches the owner's model.
