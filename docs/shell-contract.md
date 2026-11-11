# Shell contract

`internal/pascalallen/domain/shell` is the whole terminal: parser, read-only file system, command registry,
completion and history. It is pure Go (standard library only, no I/O, no goroutines, no wall clock) and builds
natively and with `GOOS=js GOARCH=wasm`. The WASM bridge, the TypeScript host and the server pre-render are thin
consumers of the API below. `internal/pascalallen/domain/content` supplies the home directory it browses.

## Construction

```go
import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

sh := shell.New(content.FS(),
	shell.WithClock(time.Now),
	shell.WithUser("visitor"),
	shell.WithHost("pascalallen.com"),
	shell.WithTheme("dark"),
)
```

| Option | Default | Notes |
|---|---|---|
| `WithClock(func() time.Time)` | Unix epoch, UTC | Only `date` reads it. The host injects the real clock. |
| `WithUser(string)` | `visitor` | Used in the prompt, `whoami`, `pwd` (`/home/<user>`). Empty is ignored. |
| `WithHost(string)` | `pascalallen.com` | Used in the prompt. Empty is ignored. |
| `WithTheme(string)` | `dark` | The starting theme, e.g. restored from storage. Unknown names are ignored. |

`content.FS() fs.FS` returns the embedded home directory. `New` accepts any `fs.FS`, so tests and other hosts may
pass an `fstest.MapFS`.

A `*Shell` holds state (working directory, previous directory, theme, history). It is not safe for concurrent use:
use one per visitor session.

## Methods

```go
func (s *Shell) Exec(line string) Result
func (s *Shell) Complete(line string, cursor int) Completion
func (s *Shell) Prompt() string
func (s *Shell) History() []string
func (s *Shell) Boot() Result
func (s *Shell) Deliver(feed string, items []FeedItem) Result
func (s *Shell) DeliverError(feed string) Result
func (s *Shell) Commands() []CommandInfo
```

- `Exec` runs one input line. It records the trimmed line in history first (blank lines are skipped, consecutive
  duplicates are collapsed, at most 200 entries are kept), then parses and runs it. It never panics; every failure
  is an error block. Work is bounded: a line over 4096 characters (code points) is rejected with one error block and
  is not recorded; a line with more than 32 sequenced commands is rejected with one error block; one `Exec` (and
  one `Deliver`) returns at most 2000 output lines, followed by a dim `output truncated after 2000 lines` line.
  `Complete` leaves a line over 4096 characters unchanged.
- `Complete` completes the word that ends at `cursor`. `cursor` and the returned `Cursor` are offsets in Unicode
  code points (runes), clamped to the line. One match returns the completed line (commands and files get a trailing
  space, directories a trailing `/`). Several matches extend the word to their longest common prefix and return
  the sorted candidates (directories with a trailing `/`). No match returns the line and cursor unchanged with no
  candidates. Command position (start of line, after `;` or `&&`) completes command names; argument position
  completes paths, directories (`cd`, `tree`), command names (`help`, `man`, `sudo`) or theme names (`theme`).
  Flags complete to nothing.
- `Prompt` is bash style without a trailing space: `visitor@pascalallen.com:~$`,
  `visitor@pascalallen.com:~/writing$`.
- `History` returns a copy, oldest first. The host walks it for the Up and Down keys.
- `Boot` returns the first screen: a banner line, the rendered `about.md`, and a row of tappable commands. It does
  not touch history.
- `Deliver` and `DeliverError` are called by the host after it handles a `fetch` effect. `feed` is `repos` or
  `packages` (`shell.FeedRepos`, `shell.FeedPackages`); anything else returns an error block.
- `Commands` lists the visible commands in registry order, for help screens and the server-rendered page. Easter
  eggs and the write commands (`sudo`, `vim`, `rm`, `mkdir`, `touch`, `mv`) work but are not listed.

## Types

```go
type Style int
const (
	StylePlain Style = iota // "plain"
	StyleBold               // "bold"
	StyleDim                // "dim"
	StyleAccent             // "accent"
	StyleSuccess            // "success"
	StyleWarning            // "warning"
	StyleError              // "error"
	StyleHeading            // "heading"
	StylePath               // "path"
)

type Span struct {
	Text  string
	Style Style
	Href  string
	Run   string
}

type Line []Span

type BlockKind int
const (
	BlockText BlockKind = iota // "text"
	BlockError                 // "error"
	BlockMedia                 // "media"
)

type Block struct {
	Kind  BlockKind
	Lines []Line
	Src   string
	Alt   string
}

type EffectKind int
const (
	EffectClear EffectKind = iota // "clear"
	EffectOpen                    // "open"
	EffectFetch                   // "fetch"
	EffectTheme                   // "theme"
)

type Effect struct {
	Kind EffectKind
	Arg  string
}

type Result struct {
	Blocks  []Block
	Effects []Effect
}

type Completion struct {
	Line       string
	Cursor     int
	Candidates []string
}

type FeedItem struct {
	Name        string
	Description string
	URL         string
	Meta        string
}

type CommandInfo struct {
	Name     string
	Synopsis string
	Summary  string
}
```

`Style`, `BlockKind` and `EffectKind` have a `String()` method returning the lower-case names in the comments
(`unknown` for out-of-range values); the bridge should serialise those names, not the integers.

Rules the host can rely on:

- A `Span` with `Href` is a link: open it in a new tab. `Href` is always `http://`, `https://` or `mailto:`.
- A `Span` with `Run` is a tappable command: on tap, echo it at the prompt and pass it to `Exec`. A span never has
  both.
- Every `Text`, `Href`, `Run`, `Src`, `Alt` and `Arg`, and every completion candidate, is valid UTF-8 (invalid
  bytes become U+FFFD) with no characters from the Unicode categories Cc (control), Cf (format: bidi overrides and
  isolates, zero-width characters, U+FEFF, soft hyphen), Zl or Zp (U+2028, U+2029). Cc, Zl and Zp become spaces;
  Cf characters are removed. The rule applies to everything that reaches a span: file content, echoed input,
  error messages that quote input, and feed items. So a `Line` never contains a newline or a direction override
  and can be rendered as one row of text. An empty `Line` is a blank row. `Completion.Line` is the visitor's own
  input and is returned as typed.
- Consecutive output of the same kind is merged into one block; text and error blocks alternate as they happen.
- Effects are applied in order, then the blocks are appended. `Exec` already discards blocks produced before a
  `clear` in the same line, so a `clear` effect means "empty the screen, then render this result".

## Parsing

Words split on spaces, tabs, `\r` and `\n`. `'...'` is literal; `"..."` honours `\"`, `\\`, `\$` and `` \` ``;
a backslash outside quotes escapes the next character. `#` at the start of a word starts a comment. `;` sequences
commands; `&&` runs the next command only if the previous one succeeded. `~`, `.` and `..` resolve inside the
home directory and can never climb above it; `/` and `/home/<user>` both mean home. A path that continues
past a regular file (`about.md/`, `about.md/..`) is `Not a directory`, and one that continues past a missing
name is `No such file or directory`. There is no globbing: `*`, `?` and `[` are literal, so pasted URLs such as
`https://example.com/?q=1` work. Pipes, redirection, background jobs, variables and command substitution are
rejected with one error line, e.g. `bash: pipes are not supported in this shell`. An unterminated quote is
``bash: unexpected EOF while looking for matching `"'``. Unknown commands are
`bash: foo: command not found`, followed by `did you mean ls?` (a `Run` span) when a visible command is within
two edits (transpositions count as one) and differs in length by at most two; words over 32 characters get no
suggestion.

## Commands

`help [command]`, `man <command>`, `ls [-la] [path]`, `cd [path|-]`, `pwd`, `cat <file>...`, `tree [path]`,
`head [-n [-]lines] <file>...` (`-n -N` drops the last N lines), `tail [-n [+]lines] <file>...` (`-n +N` starts
at line N), `grep [-i] <text> [path]` (literal, recursive), `whoami`, `echo [text]...`, `date`, `history [n]`
(entry numbers are absolute, so they keep counting after old entries are trimmed), `clear`, `open <url|file>`, `about`, `resume`, `skills`,
`experience`, `projects`, `writing`, `contact`, `education`, `repos`, `packages`, `camera`, `theme [name]` (themes:
`dark`, `light`, `amber`, `green`), `exit`; hidden: `sudo`, `vim`, `rm`, `mkdir`, `touch`, `mv` (the write
commands answer `Read-only file system`). `man <command>` documents each one.

`cat` renders files line by line: a line starting with `# ` becomes one `heading` span without the marker;
`http(s)://` URLs and email addresses become link spans; everything else is plain. There is no Markdown engine.

## Example results

Text block (`whoami`):

```go
Result{Blocks: []Block{
	{Kind: BlockText, Lines: []Line{{{Text: "visitor"}}}},
}}
```

Text block with links and tappable commands (`ls projects`, abridged):

```go
Result{Blocks: []Block{
	{Kind: BlockText, Lines: []Line{{
		{Text: "adk-agent.md", Run: "cat ~/projects/adk-agent.md"},
		{Text: "  "},
		{Text: "es-go.md", Run: "cat ~/projects/es-go.md"},
	}}},
}}
```

Error block (`cat nope`):

```go
Result{Blocks: []Block{
	{Kind: BlockError, Lines: []Line{
		{{Text: "cat: nope: No such file or directory", Style: StyleError}},
	}},
}}
```

Media block (`camera`):

```go
Result{Blocks: []Block{
	{Kind: BlockMedia, Src: "/api/v1/camera/stream", Alt: "Live camera stream from a Raspberry Pi",
		Lines: []Line{{{Text: "live from a Raspberry Pi; type clear to close it", Style: StyleDim}}}},
}}
```

Clear effect (`clear`):

```go
Result{Effects: []Effect{{Kind: EffectClear}}}
```

Open effect (`open https://github.com/pascalallen`):

```go
Result{
	Blocks: []Block{{Kind: BlockText, Lines: []Line{{
		{Text: "opening "},
		{Text: "https://github.com/pascalallen", Style: StyleAccent, Href: "https://github.com/pascalallen"},
	}}}},
	Effects: []Effect{{Kind: EffectOpen, Arg: "https://github.com/pascalallen"}},
}
```

Fetch effect (`repos`), then the host calls `/api/v1/repos` and passes the items to `Deliver`, or calls
`DeliverError` on failure:

```go
Result{
	Blocks:  []Block{{Kind: BlockText, Lines: []Line{{{Text: "fetching repositories from GitHub...", Style: StyleDim}}}}},
	Effects: []Effect{{Kind: EffectFetch, Arg: "repos"}},
}

sh.Deliver("repos", []FeedItem{{Name: "pgqueue", Description: "Postgres-backed job queue",
	URL: "https://github.com/pascalallen/pgqueue", Meta: "Go"}})
// Result{Blocks: []Block{{Kind: BlockText, Lines: []Line{
//	{{Text: "Repositories on GitHub", Style: StyleHeading}},
//	{{Text: "pgqueue", Style: StyleAccent, Href: "https://github.com/pascalallen/pgqueue"}, {Text: "  Go", Style: StyleDim}},
//	{{Text: "  Postgres-backed job queue"}},
// }}}}

sh.DeliverError("repos")
// Result{Blocks: []Block{{Kind: BlockError, Lines: []Line{
//	{{Text: "repos: could not load the list right now", Style: StyleError}},
//	{{Text: "try again: "}, {Text: "repos", Style: StyleAccent, Run: "repos"}},
// }}}}
```

Theme effect (`theme amber`); the host swaps its stylesheet and may persist the name for `WithTheme`:

```go
Result{
	Blocks:  []Block{{Kind: BlockText, Lines: []Line{{{Text: "theme set to "}, {Text: "amber", Style: StyleSuccess}}}}},
	Effects: []Effect{{Kind: EffectTheme, Arg: "amber"}},
}
```

The golden transcripts in `internal/pascalallen/domain/shell/testdata/` show the boot screen and a full session.
Regenerate them with `go test ./internal/pascalallen/domain/shell/ -run TestGolden -update`.
