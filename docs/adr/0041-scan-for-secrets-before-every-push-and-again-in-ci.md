# 0041. Scan for secrets before every push and again in CI

- **Status**: Accepted
- **Date**: 2026-11-08
- **Spec**: Owner request before publishing

## Context

The repository is about to become public, and a push publishes history, not just the latest files: a token committed and deleted later is still exposed. GitHub secret scanning only reports leaks after they are public. The owner asked for a local gate that every push must pass first.

## Decision

We will use gitleaks (MIT, single binary, about 150 built-in rules) in two places. (1) A versioned `pre-push` hook in `.githooks/`, enabled per clone with `make hooks` (`core.hooksPath`), scans exactly the commits the push would publish (`remote..local` for updates, everything no remote has for new refs) and blocks the push on any finding, printing findings redacted. It fails closed when gitleaks is missing. The only bypass is git's deliberate `--no-verify`. (2) A CI job scans the full history of every branch on each push and pull request, catching clones that never enabled the hook or bypassed it. False positives are recorded by fingerprint in `.gitleaksignore`. Before the first push the full history was scanned: 149 commits, no findings.

## Consequences

Leaks are stopped before they leave the machine, and CI backs this up. Contributors need gitleaks installed (`brew install gitleaks`) and must run `make hooks` once. Hooks are local configuration, so CI remains the enforcing gate. The hook is tested in throwaway repositories (clean push, blocked push, secret removed in a later commit, new branch, branch deletion).

## Alternatives considered

- **trufflehog**: verifies whether found credentials are live, but AGPL, heavier, and verification makes network calls.
- **detect-secrets**: requires maintaining a baseline file.
- **git-secrets**: mostly AWS patterns.
- **ggshield**: requires a GitGuardian account.
- **pre-commit hook only**: misses commits made before the hook was enabled; pre-push sees everything about to be published.
