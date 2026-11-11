# CLAUDE.md

Pointers and stable rules. Status lives in the epic issue #36 and `docs/build-log.md`, never here.

## What this is

pascalallen.com: the whole site is a bash-style terminal. A pure Go shell runs natively on the server (pre-render, tests) and as WASM in the browser. One Go binary (Gin) serves it, a small JSON API under `/api/v1` and a live MJPEG camera stream from a Raspberry Pi. Production is the Pi itself running Docker Compose (`scripts/cron-deploy` is the intended deploy path, repaired by #42); k8s manifests are an alternate path in `docs/DEPLOYING.md`.

The code under the old layout (GORM, Postgres, RabbitMQ, auth, the React SPA) is removed by #38. Do not extend it.

## Domain docs

- `GLOSSARY.md` (terms and `_Avoid_` words), `GLOSSARY-MAP.md`, `docs/adr/` (decisions), `ROADMAP.md` (order), `REVIEW.md` (review invariants), `docs/build-log.md`.
- Path-scoped rules load from `.claude/rules/`: `shell-domain.md`, `go.md`, `web-host.md`, `wasm.md`.

## Commands

Go and Node run in throwaway containers, so no running compose stack is needed to verify.

```bash
bin/verify          # the full gate, same steps as CI, from any worktree
bin/up              # build + up + compile WASM; bin/up prod = detached
bin/down            # bin/down prod likewise
bin/watch           # live reload for Go
bin/wasm            # recompile web/static/assets/wasm.{wasm,js} only
bin/exec go test ./...
bin/yarn <args>     # yarn in web/app (node container)
```

Site: http://localhost:9990. Live checks from agents use the port in their brief, never 9990.

## Architecture (target)

Module `github.com/pascalallen/pascalallen.com`; entrypoint `cmd/pascalallen`; app code under `internal/pascalallen/`.

- `domain/shell`: pure Go, stdlib only. Parser, read-only virtual filesystem, command registry, completion, history. `Exec(line)` returns blocks (styled spans, links) and effects.
- `domain/content`: `go:embed` tree that is the virtual filesystem.
- `application`: queries (`ListRepositories`, `ListPackages`), command (`RecordVisit`), ports for feeds.
- `infrastructure`: `wasm` (bridge only), `feed` (GitHub and npm clients, TTL cache), `camera` (hub), `messaging` (synchronous buses, `ctx` threaded), `container` (Wire injector), `http` (Gin routes, JSend, template).
- `web/app`: thin hand-written TypeScript host, no framework. Owns the DOM, input, rendering of blocks, scrolling, themes.
- Live commands return an effect; the host calls the API and feeds the result back. No secret ever reaches the browser.
- Layering is strict: infrastructure imports application and domain; domain imports nothing outside itself.

Layout drift is enforced by an architecture test once item 3 lands; until then `REVIEW.md` is the check.

## Generated files

- `wire_gen.go`: never hand-edit; regenerate with `go generate ./...` (denied in `.claude/settings.json`).
- `web/static/assets/` is a build output and gitignored.

## Host facts

- `compose.yaml` runs `pascalallen-go` privileged with `/dev/media0-3` and `/dev/video0` mapped and the `Dockerfile` is pinned to `linux/arm64` with libcamera and gstreamer. `bin/up` as written targets the Pi; on a Mac drop the `devices:` block or use `bin/watch`. The revamp moves the Pi bits to an overlay (#42).
- Required env is in `.env.example`. Never read or commit `.env`.

## Done means

- `bin/verify` passes on the exact tree you push, and CI is green on the PR.
- Tests are written first, use testify, and name `t.Run` cases as sentences. Each new guard has a test that fails when the guard is removed.
- No code comments; rationale goes in the PR body.
- Every shell command has help text, completion and a test; shell output is data, never HTML.
- A new CI step is added to `bin/verify` in the same PR.
- Test an empty environment and load the app in a browser after a dependency or host change.

## Workflow

GitHub issue, then branch `feature/<issue#>-<slug>`, then PR. Until the epic (#36) closes, PRs target the `revamp` integration branch and one final PR merges `revamp` into `main` (ADR 0003). The owner reviews and merges the final PR; agents never merge it. One checklist item per PR, one or two plain commits, no `Co-Authored-By` trailer. PR body: what is in and out, deviations, what was checked, `Part of #N`, last line `🤖 Generated with [Claude Code](https://claude.com/claude-code)`. Follow-ups are filed as `needs-triage` issues.

## Agents

Hub and spoke: a thin orchestrator writes briefs to `.agents/rounds/<round>/` (gitignored), launches writers in their own worktree and read-only reviewers, and owns merge order and shared files.

- `site-writer` (sonnet): one item from a brief, test-first, opens the PR.
- `site-reviewer` (opus): refute-by-default, reads `REVIEW.md`, reports only defects with a failure scenario.
- Max 4 writers and 2 reviewers at once. After two fix passes or a new defect class, stop and report.
- Never name private reference repos in code, docs, issues or PRs.
