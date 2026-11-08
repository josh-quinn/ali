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
NUL-separated pairs `name, live expansion` for each ali alias that differed from ali's definition at the first prompt (empty expansion = not an alias). NUL is used because commands may contain newlines. An empty stream still means "ran through the wrapper, nothing redefined".

## Session state in zsh (generated file, not persisted)
- `_ali_names`: array of names ali defined (unchanged).
- `_ali_prev`: associative array `name -> original alias text`, for originals ali overrode in this session.
- `_ali_expected`: associative array `name -> ali's command`, rewritten on every load.
- `_ali_overridden`: associative array filled once by `_ali_startup_check` at the first prompt.

## Doctor result
| Check | Outcomes |
|-------|----------|
| block present | ok / fail |
| block current | ok / fail (outdated) |
| block after oh-my-zsh.sh | ok / fail / skipped (no Oh My Zsh) |
| store valid | ok / fail |
| aliases not redefined later | ok / fail (lists names) / skipped (outside wrapper) |
