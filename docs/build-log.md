# Build log

A heading per epic, then one line per merged PR: `- #<n> <slug>: what landed; evidence; deviations`.

## Epic #36: terminal revamp

- claude-setup (#37): Claude Code settings, path-scoped rules, `bin/verify`, single `ci.yml`, agents, docs and ADRs 0001-0004; `bin/verify` green locally and CI on the PR; govulncheck non-blocking until #38 removes `golang-jwt/jwt`, which has no fixed version.
