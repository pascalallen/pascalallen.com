# Render contract

The DOM the browser host (`web/app/src/render.ts`, `terminal.ts`) builds from a shell `Result`, so the server
pre-render can emit the same structure and the host can adopt it without re-rendering. Result shapes are in
`docs/shell-contract.md`. Every text value reaches the DOM as text (`textContent` in the host, escaped by the server);
nothing from a `Result` is ever parsed as HTML.

## Page skeleton (`web/template/index.tmpl`)

```html
<div class="terminal" id="terminal" data-state="loading" data-wasm-src="/static/assets/shell.wasm?v=<hash>"
     data-runtime-src="/static/assets/wasm_exec.js?v=<hash>">
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
| `#terminal` | yes | `data-state` is `loading`, then `ready` or `failed` (set by the host). `data-wasm-src` and `data-runtime-src` are the module and runtime URLs from the `asset` helper (`?v=<content hash>`, cached immutably). |
| `#scroll` | yes | The scrolling region. |
| `#output` | yes | `role="log"`, `aria-live="polite"`. Holds entries only. |
| `#prompt` | yes | Text is `Shell.Prompt()`; the host replaces it after each command. |
| `#input` | yes | The real input. Do not wrap it in a `<form>` (Enter must not submit). |
| `.chips > a.chip[data-run]` | no | Tappable commands. `href="/#<command>"` so they work as deep links before the host loads. |

Head, in order: `theme.js` (a blocking script, no `defer`, under 512 bytes gzipped, that only sets `data-theme` on
`<html>` from the stored theme so a reload never paints the default theme first; it is in the shared head, so `/plain`
gets it too), the stylesheet `app.css`, and `app.js` with `defer`. There is no preload for the module: the host starts its `fetch`
at the top of app start (a preload is fetched twice by WebKit). All are same origin and go through the
`asset` helper. The host loads `wasm_exec.js` itself from `data-runtime-src` (it fetches the file, requires a
JavaScript content type, then injects a `<script>`), so a missing or HTML-served runtime puts the terminal in the
`failed` state instead of logging a syntax error. Do not add a static `wasm_exec.js` tag.

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
- **Entrance animation**: for each entry it renders (including the boot entry and results delivered later by a feed),
  the host adds `class="reveal"` to the `.entry` element and sets the custom property `--i` (via the CSSOM, to the
  zero-based index) on each of the first 24 `.block > .line` elements of that entry, counting in document order across
  the entry's blocks; the echo line is not under a `.block` and gets no `--i`. Server-rendered markup must not include
  `reveal` or `--i`: pre-rendered content must be visible without JS or animation, and the host never adds either to
  adopted `data-prerendered` entries.
- **Failure state**: when the shell cannot load, `#terminal` gets `data-state="failed"`, `#input` is disabled and the
  host appends an entry with one `block block-error` whose line reads `The terminal could not start. Read the plain
  text version.`: `s-warning` spans around `<a class="s-accent" href="/plain">plain text version</a>` (same tab, no
  `target`). The pre-rendered boot screen stays above it. In this state the host does not intercept run anchors or
  chips: it sets their `href` to `/plain`, so a tap navigates to the plain text version; host-rendered run buttons
  are disabled with a `title` that says so.
- **Modifier clicks**: a click on a run anchor with Cmd, Ctrl, Shift or Alt held, or with a non-primary button, is
  not intercepted, so the browser opens `/#<command>` as it would any link.
- **Camera**: a result with a `media` block whose `src` is `/api/v1/camera/stream` is not rendered until the host
  has read `/api/v1/camera/status`. When `data.available` is `true` the result renders as usual; otherwise (false, an
  HTTP error, a timeout or a network failure) the entry gets one `block block-error` with the `s-error` line
  `camera: the camera is offline right now`, and the stream is never requested.
  `clear` (typed by the visitor) removes every child of `#output`, pre-rendered ones included.

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
  (`s-plain gap`). They stay inline at every width, so rows of commands and `ls` columns wrap like words; no CSS stacks them.
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

`/#<encodeURIComponent(line)>`. A strict allow-list: the host runs the line on load only when it is a single command,
made only of letters, digits, spaces and `. _ / ~ @ : = , + -` (so no `;`, `&`, `|`, `<`, `>`, `$`, backtick,
backslash, quote, tab, CR or LF), whose first word is one of `help`, `man`, `ls`, `cd`, `pwd`, `cat`, `tree`, `head`,
`tail`, `grep`, `whoami`, `date`, `about`, `resume`, `skills`, `experience`, `projects`, `writing`, `contact`,
`education`. Everything else, including unknown words, `theme`, `clear`, `history`, `echo`, `open`, `repos`,
`packages`, `camera` and `exit`, is only placed in the input for the visitor to confirm. Effects (`theme`, `clear`,
`open`, `fetch`) of a link-run are ignored; they apply only to a visitor-initiated run. A visitor-initiated run
updates the hash with `history.replaceState` only for allow-listed commands; any other command removes the hash.

## Content-Security-Policy

The template carries this policy as a `<meta http-equiv>`; the header middleware (item 9) should send the same
policy plus `frame-ancestors 'none'` (not allowed in a meta tag):

```
default-src 'self'; script-src 'self' 'wasm-unsafe-eval'; style-src 'self'; img-src 'self' data:;
connect-src 'self'; font-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'
```

The host sets custom properties through the CSSOM (`style.setProperty`), which `style-src 'self'` allows; it
never writes a `style` attribute string. No inline script is needed: page config stays in `#script_config`.
