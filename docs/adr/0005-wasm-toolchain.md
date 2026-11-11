# 0005. WASM toolchain: TinyGo, checked against the standard Go build

Status: accepted

## Context

ADR 0001 left the choice between standard Go and TinyGo to a measured spike. The module must stay under a 400 KB brotli budget (409600 bytes, enforced by `bin/wasm-size` in `bin/verify` and CI). Both toolchains built the same bridge (`infrastructure/wasm`) and were measured on the same commit.

| Build | Raw | gzip -9 | brotli -q 11 | Node smoke test | Notes |
|---|---:|---:|---:|---|---|
| Go 1.26 `-ldflags='-s -w' -trimpath` | 3,327,408 | 925,336 | 699,464 | pass | over budget |
| TinyGo 0.44.0 `-opt=z -no-debug` | 485,594 | 174,859 | 135,879 | pass | chosen |
| TinyGo 0.44.0 plus `-panic=trap` | 343,959 | 141,006 | 113,198 | pass | a panic kills the runtime: rejected |
| TinyGo 0.44.0 plus `-scheduler=none` | 112,402 | n/a | n/a | fail | `select{}` deadlocks, no `resume` export: rejected |

`-panic=trap` was probed with a bridge call that panics. With the default panic handler the boundary `recover` returns an error result and the module keeps working; with `-panic=trap` the call throws `RuntimeError: unreachable` and the module is dead. The brief requires a bad input never to kill the runtime, so the default is kept.

`tinygo test ./internal/pascalallen/domain/...` does not run: every domain test binary panics at start with `reflect.Value has no flag field`. The cause is the test dependency (`testify` pulls in `go-spew`, which reads the `reflect.Value` internals that TinyGo's reflect lacks), not the shell code. So the suite cannot be the proof for TinyGo.

## Decision

Build with TinyGo 0.44.0 (`-target wasm -opt=z -no-debug`, default panic and scheduler), image pinned by digest in `bin/wasm` and the Dockerfile. The loader is the TinyGo `wasm_exec.js` from the same image, emitted next to the module.

Because `tinygo test` cannot run the domain suite, `bin/wasm-parity` replaces it. It builds the module with standard Go, replays a fixed session in Node against both (boot, every command reachable by tappable commands, every `man` page, completions, feed delivery, hostile input, the work limits) and requires byte-identical output. The 13,317-line session output matched. The standard Go build stays the reference that `go test` covers. `bin/wasm-smoke` asserts the documented shapes and bad-argument handling on the shipped module.

## Consequences

- The module is about a third of the budget (135,879 of 409,600 bytes brotli), leaving room for content growth.
- Adding an import to the bridge or domain needs a re-run of `bin/wasm` and the size check; `bin/wasm-size` fails the build if the budget is exceeded.
- TinyGo lags Go releases. A domain change that uses a standard library feature TinyGo lacks fails `bin/wasm` loudly (build error), and a behavioural difference fails `bin/wasm-parity`.
- `bin/wasm` runs `tinygo` directly when the pinned version is on PATH (the Dockerfile stage), and otherwise in the pinned image, so it needs no compose stack.
- If the parity check ever proves unmaintainable, the fallback is the standard Go build, which needs about 1.7x the budget: that would need a new budget decision.
