# Security policy

## Supported versions

ali is young and pre-1.0. Security fixes go into the latest release only; please
upgrade to the newest version before reporting.

| Version | Supported |
|---|---|
| latest release | yes |
| older releases | no |

## Reporting a vulnerability

Please **do not open a public issue** for a security problem.

Report it privately through GitHub: go to the repository's **Security** tab and
choose **Report a vulnerability**
(<https://github.com/josh-quinn/ali/security/advisories/new>). Include what you
did, what happened, the ali version (`ali --version`), and your OS and zsh version.

You can expect an acknowledgement within 7 days and an assessment within 14 days.
If the report is confirmed, a fix is released as soon as practical, with credit to
you in the release notes unless you prefer otherwise, and a GitHub security advisory
is published.

## What counts

ali writes shell code that your shell runs and edits your `.zshrc`, so these are in
scope and taken seriously:

- anything that makes ali run, or insert into a command line, text it should treat
  as data (command or shell injection through aliases, values, environments,
  imported files or the store);
- ali writing outside its config directory or `.zshrc`, following paths it should
  not, or loosening file permissions;
- secrets or other data from your aliases leaking to other users or into output;
- weaknesses in the release process (artifacts, checksums, provenance).

Out of scope: a command you deliberately stored doing what you told it to, and
problems that need an attacker who can already edit your files as you.

## How ali is hardened

Design decisions with security impact are recorded as ADRs in
[`docs/adr/`](docs/adr) (notably 0021-0025, 0036, 0040-0042). Every change runs
gosec, govulncheck, CodeQL, gitleaks over the full history, dependency review and
fuzz tests in CI, and releases carry build-provenance attestations.
