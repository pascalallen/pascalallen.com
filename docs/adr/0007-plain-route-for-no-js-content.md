# 0007. A `/plain` route carries the whole content tree without JavaScript

Status: accepted

## Context

The terminal needs WASM and JavaScript. Crawlers, text browsers and a visitor whose terminal failed to start still need the content.

## Decision

`GET /plain` (and `HEAD`) renders every file of the content tree as one HTML page through the same server renderer as the boot screen (`application/render`), indexable with its own canonical URL. The terminal's failure state links to it. The index page is the terminal with the boot screen pre-rendered; every other path is `noindex`.

## Consequences

- Content has one source (the embedded tree) and two views; `TestPlain` asserts every file and link appears.
- Leading spaces of a line become an `indent-N` class, so the text of an indented line starts at its first character.
- Shell output stays data: the renderer escapes every text value.
