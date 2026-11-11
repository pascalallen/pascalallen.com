---
paths:
  - "internal/pascalallen/domain/**"
---

# Shell and content domain

- Standard library only. No import from `application`, `infrastructure` or any module in `go.mod`, except `testify` in `_test.go` files.
- No I/O: no `os`, `net`, `io/fs` reads at run time, no clock, no randomness, no goroutines that outlive a call. Content comes from `go:embed`; the current time is a parameter.
- No `syscall/js` and no build tags. The package must compile for native and for `GOOS=js GOARCH=wasm`.
- Deterministic: the same input and state give the same `[]Block` and `Effect`s, byte for byte. Sort map iteration.
- Shell output is data (styled spans, links, effects), never HTML.
- Every command has help text, completion and a test.
