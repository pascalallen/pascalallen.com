---
paths:
  - "internal/pascalallen/infrastructure/wasm/**"
---

# WASM bridge

- Carries `//go:build js && wasm`. It is not built by `go build ./...`; item #43 adds the build script.
- Bridge only: it exposes `exec`, `complete` and `prompt` to JavaScript and converts between `syscall/js` values and domain types. No parsing, no command logic, no content.
- Return plain data (JSON-shaped objects), never HTML. Recover from panics at the boundary so a bad input cannot kill the runtime.
- The compressed binary size is under the CI budget recorded in the ADR; adding an import needs a size check.
