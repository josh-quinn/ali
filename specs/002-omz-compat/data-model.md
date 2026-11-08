# Data Model: Oh My Zsh compatibility

No change to `aliases.toml`.

## Shell definition report (wrapper to binary, not persisted)
| Form | Meaning |
|------|---------|
| `ALI_SHELL_DEF=` (empty or unset) | name is free in the shell, or the binary was run outside the wrapper |
| `ALI_SHELL_DEF=alias<TAB><text>` | name is a non-ali alias with this expansion |
| `ALI_SHELL_DEF=function` | name is a shell function |

`ALI_SHELL_DEF` is an internal interface between the wrapper and the binary of the same version; it is documented but not a public API.

## Shell state (wrapper to `ali doctor --shell-state -`, not persisted)
One line per ali-owned name: `name<TAB>live expansion` (empty expansion = not an alias in the shell).

## Session state in zsh (generated file, not persisted)
- `_ali_names`: array of names ali defined (unchanged).
- `_ali_prev`: associative array `name -> original alias text`, for originals ali overrode in this session.

## Doctor result
| Check | Outcomes |
|-------|----------|
| block present | ok / fail |
| block current | ok / fail (outdated) |
| block after oh-my-zsh.sh | ok / fail / skipped (no Oh My Zsh) |
| store valid | ok / fail |
| aliases not redefined later | ok / fail (lists names) / skipped (outside wrapper) |
