# 0001. Terminal shell in pure Go, two runtimes

Status: accepted

## Context

The site is rebuilt as a bash-style terminal. The same behaviour must run in the browser (interactive) and on the server (pre-rendered boot screen, tests, no-JS content). Writing the shell twice, in TypeScript and Go, would let the two drift.

## Decision

The whole shell lives in one pure Go package, `domain/shell`, with its content in `domain/content` (`go:embed`). It uses the standard library only, does no I/O and is deterministic, so it compiles natively and to WASM. `Exec(line)` returns blocks and effects; the host performs effects. A thin bridge in `infrastructure/wasm` exposes it to JavaScript.

## Consequences

- One source of truth: server pre-render and browser output come from the same code, and golden transcripts test both.
- The domain cannot import anything outside the standard library; an architecture test enforces it.
- The browser needs a WASM binary; its size is held to a CI budget, and the toolchain choice (standard Go or TinyGo) is decided by a measured spike in a later ADR.
- Live data (repositories, packages) arrives as effects, so the shell stays free of network code.
