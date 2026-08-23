# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Personal site + playground for pascalallen.com: a single Go binary (Gin) that serves a React SPA, a JSON API under `/api/v1`, and a live MJPEG camera stream from a Raspberry Pi. Production is the Pi itself running Docker Compose (`scripts/cron-deploy` pulls `main` and rebuilds); the k8s manifests under `internal/pascalallen/infrastructure/etc/k8s/` are an alternate deploy path documented in `docs/DEPLOYING.md`.

## Where this repo diverges from the global defaults

This predates the conventions in `~/.claude/CLAUDE.md`. Follow what's here, not the defaults, unless explicitly migrating:

- **GORM, not `database/sql`**: `database.Session` wraps `*gorm.DB`; entities carry `gorm:` struct tags; schema comes from `AutoMigrate` in `cmd/pascalallen/pascalallen.go` — there is no `migrations/` dir and no golang-migrate.
- **Async RabbitMQ buses, not synchronous in-process**: `CommandBus.Execute` publishes to the durable `commands` queue; `EventDispatcher.Dispatch` publishes to the `events` exchange. Handlers/listeners run in consumer goroutines started by `runConsumers`, so an HTTP action that executes a command returns before the handler runs. No `ctx` is threaded through handlers.
- **Wire injector lives in `cmd/pascalallen/`** (`wire.go` → `wire_gen.go`), not `infrastructure/container/`. `Container` is a plain struct; handlers/listeners are wired by struct literal in `main.go`'s `runConsumers`, not via Wire.
- **Frontend is React 18 + MobX + react-router v6 + ESLint 8 legacy `.eslintrc.json`**, no Bootstrap/TanStack. Don't upgrade toolchains incidentally.
- Known layering leak: `domain/user` imports `infrastructure/database/type.GormUlid` for its ID column. Don't spread it further.

## Commands

Everything Go runs inside the `pascalallen-go` container via `bin/exec`; everything JS runs in a throwaway `node:lts` container via `bin/yarn` (cwd `web/app`).

```bash
cp .env.example .env          # once
bin/up                        # build + up + compile WASM + tail logs; bin/up prod = detached, quiet
bin/down                      # bin/down prod likewise
bin/watch                     # postgres+adminer via compose, Go via cosmtrek/air with live reload (does NOT start RabbitMQ)
bin/wasm                      # recompile web/static/assets/wasm.{wasm,js} only

bin/exec go test ./...
bin/exec go test ./internal/pascalallen/domain/user/ -run TestRegister -v   # single package / test
bin/exec go test ./... -covermode=count -coverprofile=coverage.out
bin/exec go run github.com/google/wire/cmd/wire ./cmd/pascalallen           # regenerate wire_gen.go — never hand-edit

bin/yarn ci                   # yarn install --frozen-lockfile
bin/yarn build                # production webpack → web/static/assets/app.{js,css}
bin/yarn dev | bin/yarn watch # development build / watch
bin/yarn lint | bin/yarn lint:fix
```

Site: http://localhost:9990 · Adminer: http://localhost:8888 · RabbitMQ mgmt: http://localhost:15672.

CI (`.github/workflows/`) runs on PRs to `main`: `go mod verify`, `go build -C cmd/pascalallen`, `go test ./...`, plus `yarn install --frozen-lockfile`, a dev webpack build, and `yarn lint`. `web/static/assets/` is gitignored — it's a build output.

**Host caveat:** `compose.yaml` runs `pascalallen-go` privileged with `/dev/media0-3` and `/dev/video0` mapped, and the `Dockerfile` is pinned to `linux/arm64` with libcamera/gstreamer/rpicam packages. `bin/up` as written targets the Pi; on a dev Mac expect to drop the `devices:` block or use `bin/watch`.

## Architecture

Module `github.com/pascalallen/pascalallen.com`; all app code under `internal/pascalallen/{domain,application,infrastructure}`; entrypoint `cmd/pascalallen`.

**Startup order** (`cmd/pascalallen/pascalallen.go`): `InitializeContainer()` (Wire: GORM session, GORM repos, seeder, AMQP conn, buses) → `AutoMigrate(Permission, Role, User)` + `Seeder.Seed()` → `runConsumers` (register command handlers + event listeners, start consuming in goroutines) → `configureServer` (Gin router, `:9990`).

**Seeding** is declarative and idempotent: `infrastructure/database/seeds/auth.{permissions,roles,users}.json` are reconciled into the DB on every boot (adds/updates/removes). Edit JSON, not SQL, to change auth fixtures. IDs are ULIDs stored as `CHAR(26)`.

**Adding a command or event touches four places** — the bus deserializes by message `Type` through a hard-coded switch:
1. Struct in `application/command/` (or `application/event/`) with `CommandName()`/`EventName()` returning `reflect.TypeOf(c).Name()`; JSON tags required since it crosses RabbitMQ.
2. Handler in `application/command_handler/` (or listener in `application/listener/`) — type-asserts to the **pointer** type (`cmd.(*command.Foo)`).
3. The `switch msg.Type` in `infrastructure/messaging/command_bus.go` `processCommand` (or `event_dispatcher.go`) — without this the message is logged as unknown and dropped.
4. `RegisterHandler`/`RegisterListener` in `runConsumers`.

Existing chain to copy: `RegisterUser` → `UserRegistered` event → `UserRegistration` listener → `SendWelcomeEmail` command.

**HTTP**: `infrastructure/routes/*.go` are `Router` methods that declare groups under `/api/v1` and bind closure actions from `application/http/action/` (`HandleX(deps...) gin.HandlerFunc`). Responses use the JSend helpers in `application/http/responder/`. `middleware.AuthRequired(userRepo)` parses the `Authorization: Bearer` JWT via `infrastructure/service/tokenservice` and re-fetches the user. `NoRoute` renders `web/template/index.tmpl` so the SPA owns every non-API path; `/static` serves `web/static`.

**Camera stream**: `action/camera_hub.go` is a process-wide singleton that lazily spawns `gst-launch-1.0 libcamerasrc … jpegenc … multipartmux` on first subscriber, fans frames out to per-client buffered channels (dropping on slow consumers), and restarts gstreamer if it exits. `GET /api/v1/camera/stream` is `multipart/x-mixed-replace`; the React `CameraPage` is just an `<img src>` pointing at it.

**Server → SPA config**: `action.HandleDefault` base64-encodes `{APP_BASE_URL, APP_ENV, GITHUB_TOKEN, SLACK_DM_URL}` into a hidden `#script_config` input; `web/app/src/utilities/env.ts` decodes it. Add new runtime config in both places.

**WASM**: `infrastructure/wasm/wasm.go` (`//go:build js && wasm`) exports a `typewriter` JS global used by `index.tmpl`. It's compiled by `bin/up`/`bin/wasm`, not by `go build ./...`.

**Frontend** (`web/app/src`): entry `app.tsx`, routes in `routes/routes.tsx` guarded by `RequiresAuthentication`/`RequiresAuthorization` wrappers; MobX `RootStore` → `AuthStore` exposed through `useStore('authStore')`; all HTTP goes through `services/ApiService.ts` (axios, JSend-typed responses, JWT refresh). `@`-aliases are declared in both `tsconfig.json` and `webpack.config.js` — add new ones to both, and to the `import/order` pathGroups in `.eslintrc.json`. Prettier: 120 cols, single quotes, no trailing commas.

## Required env

See `.env.example`. The binary `log.Fatal`s at boot without reachable `DB_*` and `RABBITMQ_*` targets; `TOKEN_SECRET` signs JWTs; `GIN_MODE` selects debug/release.
