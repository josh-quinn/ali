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
NUL-separated fields: first the generation ID of the `aliases.zsh` loaded at startup, then `name, expansion` pairs for every alias the shell had at its first prompt. NUL is used because commands may contain newlines. `doctor` compares the pairs with the store only if the generation matches the current store; otherwise the check is skipped ("open a new terminal").

## Session state in zsh (generated file, not persisted)
- `_ali_names`: array of names ali defined (unchanged).
- `_ali_prev`: associative array `name -> original alias text`, for originals ali overrode in this session.
- `_ali_gen`: generation ID of the loaded definitions (hash of names and commands).
- `_ali_at_start`, `_ali_gen_at_start`: copy of the alias table and generation taken once by `_ali_startup_check` at the first prompt.

## Doctor result
| Check | Outcomes |
|-------|----------|
| block present | ok / fail |
| block current | ok / fail (outdated) |
| block after oh-my-zsh.sh | ok / fail / skipped (no Oh My Zsh) |
| store valid | ok / fail |
| aliases not redefined later | ok / fail (lists names) / skipped (outside wrapper) |
