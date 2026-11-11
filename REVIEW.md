# Review invariants

Each line names what to check and the failure it prevents. Read by `site-reviewer` and by managed code review.

- **Domain purity.** `internal/pascalallen/domain/**` imports the standard library only, does no I/O, uses no `syscall/js`, and builds for `GOOS=js GOARCH=wasm`. Prevents a shell that cannot run in the browser or cannot be tested natively.
- **Deterministic shell.** The same line and state give the same blocks. No clock, randomness or map-order dependence. Prevents flaky golden transcripts and pre-render that differs from the live shell.
- **No secret in anything served to the browser.** Page config, templates, API responses, WASM and static assets carry no token, webhook URL or credential. Prevents a repeat of the leaked GitHub token and Slack URL.
- **Shell output is data, never HTML.** File content, command arguments, the deep-link hash and feed data reach the DOM only through `createElement` and `textContent`; link targets are scheme-checked. Prevents script injection through `/#cat ...` links and feed fields.
- **Every command has help text, completion and a test.** `help`, `man` and tab completion list it; a table test covers it. Prevents commands that exist but cannot be discovered.
- **No content loss.** Every item of the content source list (resume sections, all 18 publications plus the SABJ piece, social links, packages, projects) is present in the content tree. Prevents the rebuild silently dropping the old page's content.
- **No-JS page carries the full content.** The pre-rendered HTML holds every content section, meta and OG tags, and JSON-LD. Prevents an empty page for crawlers and script-blocked visitors.
- **Reduced motion and keyboard access.** Animation respects `prefers-reduced-motion`; every action works from the keyboard; output sits in an `aria-live` region. Prevents an inaccessible terminal.
- **WASM size budget.** The compressed binary stays under the CI budget; a new import that grows it is justified. Prevents a slow first render.
- **Layering.** `application` never imports `infrastructure`; Wire injectors live in `infrastructure/container`; `wire_gen.go` is generated. Prevents dependency drift.
- **Errors.** Failures are wrapped with `%w` and reach the JSend responder; none is swallowed; not-found reads return `nil, nil`.
- **Feeds and visits.** Outbound calls take a context with a timeout; results are cached with a TTL; the visit endpoint is rate-limited. Prevents a slow upstream or a visitor flood from stalling the site.
- **No code comments.** Rationale belongs in the PR body.
- **Tests fail when their guard is removed.** For each new guard, a test must break if the guard is deleted. A test that still passes is a defect.
- **Gate parity.** A new CI step appears in `bin/verify` in the same PR.
