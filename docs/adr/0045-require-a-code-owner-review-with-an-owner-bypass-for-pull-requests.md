# 0045. Require a code-owner review, with an owner bypass for pull requests

- **Status**: Accepted
- **Date**: 2026-11-09
- **Spec**: Owner request after the Scorecard run of 2026-11-09; reverses one alternative rejected in ADR 0043

## Context

ADR 0043 protected `main` without requiring an approving review, because a solo maintainer cannot approve their own pull requests and every merge would need an admin bypass. The owner has decided the higher OpenSSF Scorecard Branch-Protection score is worth that bypass, and it also means any future contributor's pull request cannot be merged without the owner's review.

## Decision

We will require one approving review from a code owner (`.github/CODEOWNERS`: the owner) on every pull request to `main`, in addition to the rules of ADR 0043 (no deletion, no force-push, pull requests only, required checks). Repository admins may bypass these rules **only when merging a pull request** (ruleset bypass mode `pull_request`), never by pushing directly. The owner merges their own pull requests with `gh pr merge --admin --merge` only after every required check has passed; the bypass is not used to skip failing checks.

## Consequences

Outside contributions now need the owner's approval. The owner's own merges, including Dependabot's, use the admin bypass, so Scorecard still records them as unreviewed (Code-Review) and partly discounts the bypass (Branch-Protection). Checks passing before an admin merge is now a matter of discipline rather than enforcement for the owner. When the project gains a second maintainer, add them to CODEOWNERS and drop the bypass.

## Alternatives considered

- **Keep no required review (ADR 0043)**: lower Branch-Protection score and no enforced review of outside contributions.
- **Bypass mode `always`**: would also allow direct pushes to `main`.
- **A second account approving the owner's pull requests**: games Code-Review without adding a real reviewer.
