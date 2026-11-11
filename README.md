# pascalallen.com

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pascalallen/pascalallen.com)
![GitHub](https://img.shields.io/github/license/pascalallen/pascalallen.com)

![Screenshot of the terminal](docs/screenshot.png)

The personal site of Pascal Allen, written as a bash-style terminal. Visitors list, read and search a read-only home
directory (`ls`, `cd`, `cat`, `grep`, `about`, `resume`, `projects`, `writing`), see live GitHub and npm lists, and
watch a live camera stream from a Raspberry Pi. Every page works without JavaScript at `/plain`.

## Run locally

Requires Docker with Compose.

```bash
cp .env.example .env
bin/up
```

The site is at http://localhost:9990/. `bin/up` builds the image and follows the logs; `bin/down` stops it. On a
Raspberry Pi with a camera it picks the Pi overlay (`compose.pi.yaml`) by itself; force one with `bin/up base`,
`bin/up pi` or `bin/up prod`. Required settings are listed in `.env.example`.

## Test

```bash
bin/verify
```

Runs the same steps as CI in throwaway containers: build, gofmt, vet, `go test -race -cover`, `go mod tidy` and Wire
drift, govulncheck, the comment check, frontend lint, tests and build, the WASM build with its size, smoke and parity
checks, the Playwright e2e suite (`bin/e2e`), shellcheck and the deploy script tests. `bin/yarn <args>` runs yarn in
`web/app`.

## Deploy

Production is a Raspberry Pi running Docker Compose, redeployed from `main` by `scripts/cron-deploy`. See
[docs/DEPLOYING.md](docs/DEPLOYING.md).

## Architecture

- `internal/pascalallen/domain/shell`: a pure Go shell (parser, read-only virtual filesystem, commands, completion).
  `Exec(line)` returns styled blocks and effects, never HTML.
- `internal/pascalallen/domain/content`: the embedded content tree that is the virtual filesystem.
- `internal/pascalallen/infrastructure/wasm`: the bridge that exposes the same shell to JavaScript (TinyGo, ADR 0005).
- `web/app`: a small hand-written TypeScript host that owns the DOM, input and rendering; no framework.
- `internal/pascalallen/application/render`: server-side HTML for the boot screen, `/plain` and the SEO head.
- `internal/pascalallen/infrastructure/feed`: GitHub and npm clients with a TTL cache; secrets never reach the browser.
- `internal/pascalallen/infrastructure/camera`: the camera hub (lazy start, grace period, 503 when offline).
- `internal/pascalallen/infrastructure/routes`: Gin routes, security headers, hash-versioned static assets.
- `cmd/pascalallen`: entrypoint; Wire wires the container.
- Decisions are in `docs/adr/`; the terms in `GLOSSARY.md`.

## Contributing

Pull requests are welcome; open an issue first for major changes. `bin/verify` must pass.

## License

[MIT](LICENSE)
