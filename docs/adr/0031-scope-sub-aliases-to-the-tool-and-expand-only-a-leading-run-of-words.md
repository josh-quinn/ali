# 0031. Scope sub-aliases to the tool and expand only a leading run of words

- **Status**: Proposed
- **Date**: 2026-11-08
- **Spec**: specs/003-command-expansion FR-001-FR-003

## Context

The owner wants `g` after `k` (kubectl) to differ from `g` after `h` (helm), with shared words usable anywhere and tool words winning. Expanding any matching word anywhere would rewrite real arguments (a pod named `p`).

## Decision

We will key sub-aliases by the tool the line runs (first word after alias expansion and wrappers, as in ADR 0008), so every alias for kubectl shares kubectl's sub-aliases. After the tool, consecutive sub-alias words expand (tool's own, then shared); the first word that is not a sub-alias stops expansion for that command; `\word` and quoted words stay literal; expansions are not re-expanded.

## Consequences

Predictable, no surprises in arguments. Sub-aliases after a real argument (`k logs mypod p`) do not expand; users put them first or type the full word.

## Alternatives considered

- **Key by alias name (`k`)**: `kubectl g p` and other kubectl aliases would not share words.
- **Expand anywhere**: corrupts arguments.
- **Nested tree (verb then noun)**: more to define; flat per-tool scope matches the owner's model.
