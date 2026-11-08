# Data Model: Command expansion

## Store (`aliases.toml`, synced, human-editable)

```toml
version = 2            # 1 when none of the tables below exist (research R12)

[kubectl]              # aliases by group, unchanged from feature 001
k = "kubectl"
kgp = "kubectl get pods -n {{ns}}"

[_subs.kubectl]        # sub-aliases scoped to tool "kubectl"
g = "get"
p = "pods"
n = "-n {{ns}}"

[_subs.helm]
g = "get values"

[_shared]              # sub-aliases usable after any tool (tool's own wins)
o = "-o wide"

[_vars.ns]
default = "one"        # optional; must be a key in values
[_vars.ns.values]
one = "namespace-1"
sys = "kube-system"

[_envs.dev]
[_envs.dev.set]
ns = "one"             # a value key of the variable, or a literal
ctx = "dev-cluster"

[_envs.prod]
protected = true
[_envs.prod.set]
ns = "prod"
```

### Entities and rules
| Entity | Fields | Rules |
|--------|--------|-------|
| Sub-alias | tool, name, expansion | name follows alias-name rules (ADR 0021 reserved words excluded) and must not start with `@` or `\`; tool is a group-style name (`^[a-z0-9][a-z0-9_-]*$`); expansion non-empty, placeholders outside quotes; unique per tool |
| Shared sub-alias | name, expansion | as above; a tool sub-alias of the same name shadows it for that tool |
| Variable | name, values (key → text), default | name `^[a-z][a-z0-9_]*$`, not `env`; keys `^[a-z0-9][a-z0-9_-]*$` (so `@kube-system` works); default must be a key; values non-empty, no NUL |
| Environment | name, protected, set (variable → key or literal) | name `^[a-z][a-z0-9_-]*$`; every variable in `set` must exist; a value matching one of the variable's keys means that key |
| Placeholder | `{{name}}` in an alias command or sub-alias expansion | recognised only if `name` is a variable; must not be inside quotes; deleting/renaming a referenced variable is refused |

## Machine-local files (not synced)
| File | Content |
|------|---------|
| `settings.toml` | `expand = "inline"` (default) or `"runtime"`; `prompt = "right"` (default) or `"off"` |
| `global-env` | one line: the global environment name, or empty |

## Shell session state (generated `aliases.zsh`, not persisted)
| Name | Purpose |
|------|---------|
| `_ali_trigger` | assoc: first words that can start an expandable line (R2) |
| `_ali_subs_<tool>` | arrays of sub-alias names per tool, plus `_ali_shared`, for the Space widget gate |
| `_ali_protected` | assoc: protected environment names (prompt colour) |
| `ALI_ENV` | exported per-terminal environment (set by the wrapper) |

## Resolution order (FR-009)
explicit `@key` / `@var=value` → `@@env` (line) → `ALI_ENV` (terminal) → `global-env` → variable default → error.
