# 0010. Static assets are versioned by content hash and cached immutably

Status: accepted

## Context

The WASM module and the bundle change on deploy, and stale copies in a browser or CDN break the module and runtime pairing. A fixed filename cannot be cached for long, and long caching needs a name that changes with content.

## Decision

The server hashes each asset (SHA-256, first 12 hex characters) and templates its URL as `/static/assets/<name>?v=<hash>`. A request whose `v` equals the file's current hash is served `Cache-Control: public, max-age=31536000, immutable`; anything else is `no-cache` with an `ETag` derived from the hash. `yarn build` writes `.br` and `.gz` siblings (`compress.mjs`) and the handler serves the best accepted one with `Vary: Accept-Encoding`, ignoring a sibling older than its source file so a dev rebuild cannot serve stale bytes. The dev webpack config deletes stale siblings on emit.

## Consequences

- A deploy changes the hash and so the URL; nothing needs purging.
- The hash is computed per file and remembered by size and modification time.
- Only `yarn build` produces siblings; without them the handler serves the file as is.
