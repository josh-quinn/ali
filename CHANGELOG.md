# Changelog

## Unreleased
- Security: reserve names the shell integration relies on and escape its builtins (ADR 0021); private config permissions (0022); write through symlinks (0023); fail instead of using relative paths without HOME (0024); escape control characters in output and reject NUL (0025); importer skips unquoted shell metacharacters; CI hardened and Dependabot enabled (0026).
- Initial implementation: `add`, `ls`, `groups`, `search`, `update`, `regroup`, `rm`, `init zsh`, `import`.
