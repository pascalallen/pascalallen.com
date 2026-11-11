---
paths:
  - "cmd/**/*.go"
  - "internal/**/*.go"
---

# Go conventions

- Layering is one-directional: `domain` imports nothing outside itself, `application` imports `domain`, `infrastructure` imports both. An `application` to `infrastructure` import is a bug to fix, never an exception to add.
- Commands, queries and events go through the synchronous in-process buses behind the `messaging` interfaces. Thread `ctx` through every handler. Handlers are structs with exported deps, registered by struct literal in `main.go`.
- The Wire injector lives in `infrastructure/container` (`wire.go`, generated `wire_gen.go`, never hand-edited). Regenerate with `go generate ./...`.
- HTTP uses Gin, closure actions (`HandleX(deps) gin.HandlerFunc`) and JSend responders. Routes declare their groups under `/api/v1`.
- Not-found reads return `nil, nil`; the handler decides. Wrap failures with `%w` only.
- Nothing secret reaches the browser: runtime config injected into the page holds public values only.
- No code comments. Tests use testify with `t.Run` names that read as sentences.
