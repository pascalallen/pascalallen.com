# 0003. Integration branch `revamp`

Status: accepted

## Context

The rebuild deletes most of the existing code and replaces it over about eleven PRs. Production deploys from `main`, so landing a half-built site there would break it.

## Decision

All item PRs target a long-lived `revamp` branch. CI runs on pull requests to `main` and `revamp`. When the epic is verified end to end, one final PR merges `revamp` into `main`, reviewed and merged by the owner. Items 1 and 2 land alone because they touch shared files.

## Consequences

- `main` and the Pi keep serving the current site until the final merge.
- `revamp` can drift from `main`; unrelated fixes on `main` are merged into `revamp` before the final PR.
- Agents may merge into `revamp` once gates pass; they never merge into `main`.
- When the epic closes, this ADR is superseded by plain `main`-based flow.
