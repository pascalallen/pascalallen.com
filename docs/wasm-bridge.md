# WASM bridge

`internal/pascalallen/infrastructure/wasm` compiles the shell (`docs/shell-contract.md`) to `shell.wasm` and exposes
it to JavaScript as one global object, `window.__shell`. The bridge holds no logic: it validates arguments, calls
the shell and turns the result into plain JavaScript data. The TypeScript host renders that data and never receives
HTML.

## Loading

1. Load `/static/assets/wasm_exec.js` (from the same toolchain as the module), then instantiate
   `/static/assets/shell.wasm` with `new Go()` and `go.run(instance)`.
2. `go.run` never resolves while the shell lives: the Go program registers `window.__shell`, then calls
   `window.__shellReady()` if the host defined it, then blocks forever. Define `window.__shellReady` before running
   the module; it is called exactly once, after `window.__shell` is complete. If it throws, the throw is swallowed
   and `console.error('shell: __shellReady threw')` is logged once; `window.__shell` stays usable.
3. The server sends `Content-Type: application/wasm` and, when `Accept-Encoding` allows, a precompressed `.br` or
   `.gz` sibling (with `Vary: Accept-Encoding`). Every file under `web/static` has a content hash computed once at
   startup. A request is cached as `immutable` for a year only when its `?v=` equals the hash of the file served;
   anything else gets `no-cache` and an ETag derived from the hash, so revalidation is cheap. The page template has
   an `asset` function: `{{ asset "shell.wasm" }}` renders `/static/assets/shell.wasm?v=<hash>` (without the query if
   the file is absent), and `.wasmVersion` holds the module hash (empty when absent). The page template passes both
   URLs to the host as `data-wasm-src` and `data-runtime-src` on `#terminal`. The page does not preload the module: a
   `<link rel="preload" as="fetch">` is requested a second time by WebKit, which does not reuse it for the host's
   `fetch`, so the module would transfer twice. Instead `loadShell` starts the module `fetch` synchronously when it is
   called (the top of app start), in parallel with the runtime script request, and streams that single response into
   `WebAssembly.instantiateStreaming`; when streaming is refused (a wrong content type or no support) the same response is read as a buffer, so the module is never requested twice.
   The e2e suite asserts exactly one `shell.wasm` request per page load in Chromium and WebKit.
One `Shell` exists per module instance, built over the embedded content with the real clock, user `visitor`, host
`pascalallen.com` and the `dark` theme. A `theme` command returns a `theme` effect that the host applies and persists
itself; on ready the host calls `setTheme` with the stored theme so the shell's `theme` listing agrees with the page.

## Functions

| Call | Returns |
|---|---|
| `__shell.boot()` | Result |
| `__shell.exec(line: string)` | Result |
| `__shell.complete(line: string, cursor: number)` | Completion |
| `__shell.prompt()` | string |
| `__shell.history()` | string[] |
| `__shell.commands()` | CommandInfo[] |
| `__shell.deliver(feed: string, items: FeedItem[])` | Result |
| `__shell.deliverError(feed: string)` | Result |
| `__shell.setTheme(name: string)` | boolean: true when the theme was set, false for an unknown name |

`cursor` is a code point offset. It must be an integer number from 0 to 2147483647; a fractional, negative,
out-of-range, `NaN` or non-number cursor is rejected with an error Result (it is not truncated or clamped here; the
shell clamps offsets beyond the line).

## Shapes

All kinds and styles are the lowercase names from the contract. Every key is always present; absent strings are `""`.

```ts
type Span = { text: string; style: 'plain'|'bold'|'dim'|'accent'|'success'|'warning'|'error'|'heading'|'path'; href: string; run: string };
type Block = { kind: 'text'|'error'|'media'; lines: Span[][]; src: string; alt: string };
type Effect = { kind: 'clear'|'open'|'fetch'|'theme'; arg: string };
type Result = { blocks: Block[]; effects: Effect[] };
type Completion = { line: string; cursor: number; candidates: string[] };
type FeedItem = { name: string; description: string; url: string; meta: string };
type CommandInfo = { name: string; synopsis: string; summary: string };
```

`blocks`, `effects`, `lines` and `candidates` are always arrays, never `null`. An empty line is an empty array.

## Errors

A call never throws and never kills the runtime: argument conversion and the shell call both run behind a recover
boundary. A call with the wrong number or types of arguments, a hostile value, or an internal panic returns an error
Result instead of the documented value:

```js
{ blocks: [{ kind: 'error', lines: [[{ text: 'shell: exec expects 1 string argument', style: 'error', href: '', run: '' }]], src: '', alt: '' }], effects: [] }
```

For `complete`, `prompt`, `history`, `commands` and `setTheme` the host can detect this by shape (`blocks` present). The messages
are exactly:

| Case | Text |
|---|---|
| `boot`, `prompt`, `history`, `commands` with arguments | `shell: <name> expects no arguments` |
| `exec` without exactly one string | `shell: exec expects 1 string argument` |
| `complete` without a string and a number | `shell: complete expects a string and a number` |
| `complete` with a bad cursor | `shell: complete expects a non-negative integer cursor` |
| `deliver` without a feed string and an array | `shell: deliver expects a feed name and an array of items` |
| `deliver` with a bad item | `shell: deliver items must be objects with string name, description, url and meta` |
| `deliverError` without exactly one string | `shell: deliverError expects 1 string argument` |
| `setTheme` without exactly one string | `shell: setTheme expects 1 string argument` |
| a panic or a throwing JavaScript value | `shell: <name> failed` |

An unknown feed name is not a bridge error: the shell answers with its own error block.

## Argument conversion

Strings, numbers and booleans are read as is. Arrays and objects are read to a depth of 4 (deeper values become
null). Objects must be plain (prototype `Object.prototype` or `null`) and every own enumerable property must be a
data property: getters, class instances, `Date`, `Map` and similar make the value invalid, and Proxies with throwing
traps are caught. Feed items need `name`, `description`, `url` and `meta` as strings; other plain data keys are
ignored.

## Host loop

```js
const r = __shell.exec(line);
for (const e of r.effects) { /* clear, open, theme: apply; fetch: call the API, then deliver / deliverError */ }
render(r.blocks);
```
