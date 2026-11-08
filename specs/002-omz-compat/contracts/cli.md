# CLI Contract: Oh My Zsh compatibility

| Command | Change |
|---------|--------|
| `ali add <name> <command>` | Additionally warns on stderr: `warning: <name> is already an alias in this shell (<text>); ali's alias will take precedence` or `warning: <name> is a shell function; ali's alias will take precedence`. Exit code unchanged. |
| `ali rm`, `ali update` | No output change; live-shell restoration happens in the wrapper. |
| `ali doctor` | New. Prints one line per check: `ok`, `FAIL` or `skipped`, with a fix hint for failures. Exit 0 if no check failed, 1 otherwise. Hidden flag `--shell-state -` is used by the wrapper. |
| `ali init zsh` | New outcome: `updated the ali block in <rc>` with `backup: <path>` when the block is outdated. Unbalanced markers: error, exit 1, file unchanged. |
