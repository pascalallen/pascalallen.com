# Glossary

Terms of the terminal site. Use these words; avoid the listed alternatives.

**Shell**:
The pure Go package that parses a line, runs a command against the virtual filesystem and returns blocks and effects. It runs natively and as WASM.
_Avoid_: terminal engine, interpreter, console, REPL

**Command**:
A named shell operation (`ls`, `cat`, `repos`) with help text, completion and a test.
_Avoid_: action, handler, script, builtin

**Block**:
One unit of shell output: a list of styled spans, optionally with links.
_Avoid_: line, message, response, HTML

**Span**:
A run of text inside a block with one style and at most one link target.
_Avoid_: token, fragment, element

**Effect**:
A request the shell returns for work it cannot do itself (such as `fetch repos`); the host performs it and feeds the result back.
_Avoid_: side effect, callback, event, job

**Virtual filesystem**:
The read-only tree the shell navigates with `ls`, `cd` and `cat`, backed by the embedded content.
_Avoid_: file system, storage, CMS

**Content**:
The embedded markdown tree (about, resume, contact, experience, projects, writing) that is the virtual filesystem.
_Avoid_: data, posts, pages, copy

**Host**:
The thin TypeScript layer in `web/app` that owns the DOM, input, rendering of blocks, scrolling and themes.
_Avoid_: frontend, UI, client app, SPA

**Bridge**:
The `js && wasm` Go package that exposes `exec`, `complete` and `prompt` to the host and holds no logic.
_Avoid_: wrapper, adapter, glue, API

**Feed**:
A server-side client for public data (GitHub repositories, npm packages) with an in-memory TTL cache.
_Avoid_: scraper, integration, sync

**Pre-render**:
The boot screen and content HTML produced on the server by running the same shell package, so content shows before WASM arrives.
_Avoid_: SSR, static build, snapshot

**Boot screen**:
The first output shown on load: the prompt and the opening commands' blocks.
_Avoid_: splash, landing page, intro

**Camera hub**:
The process-wide singleton that spawns gstreamer on the first subscriber and fans MJPEG frames out to viewers.
_Avoid_: stream service, webcam manager, broadcaster
