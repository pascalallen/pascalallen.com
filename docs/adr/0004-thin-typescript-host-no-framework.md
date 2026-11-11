# 0004. Thin TypeScript host, no framework

Status: accepted

## Context

The current frontend is React 18 with MobX and a router for what is now a single screen. The shell in Go (ADR 0001) already owns parsing, content and output, so the browser only has to render blocks, handle input and manage the viewport.

## Decision

`web/app` is a hand-written TypeScript host with no UI framework, router or state library. It owns the DOM, keyboard and mobile input, rendering of blocks, scrolling and themes. It builds DOM with `createElement` and `textContent` and never injects shell output as HTML. The toolchain stays Webpack 5, with ESLint 9 flat config and Prettier (120 columns, single quotes).

## Consequences

- A small bundle and no framework upgrade treadmill.
- Rendering and focus management are written by hand and covered by browser e2e tests.
- Output safety rests on one rule, checked in review: no `innerHTML` with shell output, content, arguments, the URL hash or feed data.
- Accessibility (`aria-live` output, keyboard access, reduced motion) is the host's job, not a library's.
