---
name: verify
description: Run the full verification gate (the same steps as CI) from any worktree and report pass or fail per step. Use before claiming work is done, before committing, and before opening a PR.
allowed-tools: Bash(bin/verify) Bash(git diff *) Bash(git status *) Bash(docker info *)
---

# Verify

Run `bin/verify` from the worktree root. It is the single source of truth for the gate: it runs the same steps as `.github/workflows/ci.yml`, in the same order, in throwaway containers pinned to the CI Go and Node versions, and stops at the first failure. It needs Docker running, not the compose stack.

If `docker info` fails, report "not run (Docker not running)", not a pass.

Report each step as pass or fail and show the full output of the first failure. A Wire drift failure means the committed `wire_gen.go` was stale: leave the regenerated file for the author to commit.

When an item adds a CI step, it adds the same step to `bin/verify` in the same PR.
