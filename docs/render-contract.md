# Render contract

The DOM the browser host (`web/app/src/render.ts`, `terminal.ts`) builds from a shell `Result`, so the server
pre-render can emit the same structure and the host can adopt it without re-rendering. Result shapes are in
`docs/shell-contract.md`. Every text value reaches the DOM as text (`textContent` in the host, escaped by the server);
nothing from a `Result` is ever parsed as HTML.

## Page skeleton (`web/template/index.tmpl`)

```html
<div class="terminal" id="terminal" data-state="loading" data-wasm-src="/static/assets/shell.wasm">
  <div class="scroll" id="scroll">
    <div class="output" id="output" role="log" aria-live="polite" aria-label="Terminal output">
      <!-- entries -->
    </div>
    <div class="prompt-line">
      <label class="prompt" id="prompt" for="input">visitor@pascalallen.com:~$</label>
      <input class="input" id="input" type="text" name="command" autocapitalize="off" autocomplete="off"
             autocorrect="off" spellcheck="false" enterkeyhint="go">
    </div>
  </div>
  <nav class="chips" aria-label="Commands">
    <a class="chip" href="/#help" data-run="help">help</a>
    <!-- one chip per command -->
  </nav>
</div>
```

| Element | Required | Notes |
|---|---|---|
| `#terminal` | yes | `data-state` is `loading`, then `ready` or `failed` (set by the host). `data-wasm-src` is the module URL; append `?v=<hash>` for cache busting. |
| `#scroll` | yes | The scrolling region. |
| `#output` | yes | `role="log"`, `aria-live="polite"`. Holds entries only. |
| `#prompt` | yes | Text is `Shell.Prompt()`; the host replaces it after each command. |
| `#input` | yes | The real input. Do not wrap it in a `<form>` (Enter must not submit). |
| `.chips > a.chip[data-run]` | no | Tappable commands. `href="/#<command>"` so they work as deep links before the host loads. |

Scripts: `wasm_exec.js` then `app.js`, both `defer`, both same origin; stylesheet `app.css`.

## Entries

Each command is one entry. The boot screen is an entry with no echo line.

```html
<div class="entry entry-boot" data-prerendered="">   <!-- server-rendered boot screen -->
  <div class="block block-text">...</div>
</div>

<div class="entry">                                  <!-- one executed command -->
  <div class="line echo"><span class="prompt">visitor@pascalallen.com:~$</span> <span class="cmd">ls</span></div>
  <div class="block block-text">...</div>
</div>
```

- **`data-prerendered`**: on load, if `#output` contains any element with `data-prerendered`, the host keeps the
  existing children and does not call `boot()`. Otherwise it renders `boot()` into `<div class="entry entry-boot">`
  without the attribute. Put the attribute on the boot entry, not on `#output`.
- The echo line is one text node `" "` between `.prompt` and `.cmd`.
- The host adds `class="reveal"` and an inline `--i` custom property to lines it renders for the entrance
  animation; the server must not emit either.

## Blocks, lines, spans

| Result | Element | Classes / attributes |
|---|---|---|
| `Block{Kind: text}` | `<div>` | `block block-text` |
| `Block{Kind: error}` | `<div>` | `block block-error` |
| `Block{Kind: media}` | `<div>` | `block block-media`; first child `<img class="media" src alt>` when the source is allowed (below), then its lines |
| unknown kind | `<div>` | `block block-text` |
| `Line` | `<div>` | `line`; an empty `Line` is an empty `<div class="line"></div>` |
| plain span | `<span>` | `s-<style>` |
| link span (`Href` allowed) | `<a>` | `s-<style>`, `href=<normalised URL>`, `target="_blank"`, `rel="noopener noreferrer"` |
| link span (`Href` refused) | `<span>` | `s-<style>`; text kept, no link |
| run span | host: `<button type="button">`; server: `<a href="/#<encodeURIComponent(run)>">` | `run s-<style>`, `data-run=<command>` |

- Styles: `plain`, `bold`, `dim`, `accent`, `success`, `warning`, `error`, `heading`, `path`. Anything else
  becomes `s-plain`. A span with both `Href` and `Run` cannot happen; the host checks `Href` first.
- **Gap spans**: a span whose text is two or more spaces and nothing else gets an extra class `gap`
  (`s-plain gap`). On narrow screens CSS turns gaps into row breaks so column output stacks.
- Classes are emitted in exactly the order shown (`run s-accent`, `s-plain gap`).
- **Link schemes**: `http:`, `https:`, `mailto:` only, checked by parsing the URL (the host uses `new URL(href)`,
  no base, and writes `url.href` back, so `https://example.com` becomes `https://example.com/`). The server
  should normalise the same way (Go `url.Parse` then `String()` differs on the trailing slash; either emit
  `url.href`-equivalent output or accept the host's rendering of later commands differs only in that slash).
- **Media source**: resolved against the page origin; allowed only when the result has the same origin. The
  host writes the path plus query. Refused media renders the block's lines without an `<img>`.
- **Run spans** work without JS as deep links. With JS, the host handles clicks on any `[data-run]` element
  inside `#terminal` (button or anchor), cancels navigation and runs the command, so server anchors need no
  upgrade step.

## Deep links

`/#<encodeURIComponent(line)>`. The host runs the line on load only when its first word is in `commands()` minus
`open`, `repos`, `packages`, `camera`, and the line contains none of `;`, `&`, `|`, `<`, `>`, `$`, a backtick, a
backslash, CR or LF. Anything
else is placed in the input for the visitor to confirm. A run updates the hash with `history.replaceState`.

## Content-Security-Policy

The template carries this policy as a `<meta http-equiv>`; the header middleware (item 9) should send the same
policy plus `frame-ancestors 'none'` (not allowed in a meta tag):

```
default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self'; img-src 'self' data:;
connect-src 'self'; font-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'
```

The host sets custom properties through the CSSOM (`style.setProperty`), which `style-src 'self'` allows; it
never writes a `style` attribute string. No inline script is needed: page config stays in `#script_config`.
