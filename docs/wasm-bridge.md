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
   the module; it is called exactly once, after `window.__shell` is complete.
3. The page is served with a `?v=<content hash>` query on the module URL so it can be cached as immutable. The hash
   is exposed to the template as `.wasmVersion` (empty when the module is absent). The server sends
   `Content-Type: application/wasm` and, when `Accept-Encoding` allows, a precompressed `.br` or `.gz` sibling.

One `Shell` exists per module instance, built over the embedded content with the real clock, user `visitor`, host
`pascalallen.com` and the `dark` theme. There is no theme setter: the contract has theme only as a construction
option, and a `theme` command returns a `theme` effect that the host applies and persists itself.

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

`cursor` is a code point offset and must be a finite number (it is truncated to an integer).

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

A call never throws and never kills the runtime. A call with the wrong number or types of arguments (or an unknown
feed item shape), and any internal panic, returns an error Result instead of the documented value:

```js
{ blocks: [{ kind: 'error', lines: [[{ text: 'shell: exec expects 1 argument', style: 'error', href: '', run: '' }]], src: '', alt: '' }], effects: [] }
```

For `complete`, `prompt`, `history` and `commands` the host can detect this by shape (`blocks` present). Feed
items need `name`, `description`, `url` and `meta` as strings; other keys are ignored.

## Host loop

```js
const r = __shell.exec(line);
for (const e of r.effects) { /* clear, open, theme: apply; fetch: call the API, then deliver / deliverError */ }
render(r.blocks);
```
