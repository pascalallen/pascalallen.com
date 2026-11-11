# Build log

A heading per epic, then one line per merged PR: `- #<n> <slug>: what landed; evidence; deviations`.

## Epic #36: terminal revamp

- #48 claude-setup: Claude Code settings, path-scoped rules, `bin/verify`, single `ci.yml`, agents, docs and ADRs 0001-0004; `bin/verify` green locally and CI on the PR.
- #49 shell-core: the shell and content as a pure Go domain package (parser, virtual filesystem, commands, completion, history) with golden tests and a stdlib-only architecture test.
- #50 demolition: removed auth, database, broker, the React SPA and the old WASM; ADR 0002.
- #51 camera: camera hub moved to infrastructure with an injected source, 503 when unavailable and a status endpoint.
- #52 feeds: repository and package feeds with a TTL cache, visit ping, synchronous buses, Wire container.
- #53 deploy: multi-stage image, Pi compose overlay, safe `scripts/cron-deploy` with tests.
- #54 host-ui: thin TypeScript terminal host, themes, chips, mobile layout; ADR 0004.
- #55 prerender-seo: server pre-render of the boot screen and `/plain`, SEO head, security headers.
- #56 wasm-bridge: the shell exposed to JavaScript as a TinyGo WASM module with size budget, smoke and parity checks; ADR 0005.
- #57 live-e2e: live commands end to end and the Playwright e2e suite in `bin/verify` and CI.
- #58 sweep: dead code and the Kubernetes deploy path removed (ADR 0006), `go mod tidy` stable and checked, docs corrected, ADRs 0007-0010, stale precompressed siblings ignored, prose measure and hanging indent fixes, checks for comments, routes and `CLAUDE.md`.
- Final verification on the `revamp` head (2026-11-11, base compose on a Mac, Chrome): transfer sizes with brotli are `shell.wasm` 137,611 bytes, `app.js` 4,749, `wasm_exec.js` 4,368, `app.css` 2,189, `theme.js` 205 and the page 6,991, about 156 KB for a first load; first-render timing was not measured.
