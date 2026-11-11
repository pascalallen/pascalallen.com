package plain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

func newBridge() *Bridge {
	return New(shell.New(content.FS(), shell.WithClock(func() time.Time { return time.Unix(0, 0) })))
}

func asMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	require.True(t, ok, "expected a map, got %T", v)
	return m
}

func asSlice(t *testing.T, v any) []any {
	t.Helper()
	s, ok := v.([]any)
	require.True(t, ok, "expected a slice, got %T", v)
	return s
}

func firstText(t *testing.T, res any) string {
	t.Helper()
	blocks := asSlice(t, asMap(t, res)["blocks"])
	require.NotEmpty(t, blocks)
	lines := asSlice(t, asMap(t, blocks[0])["lines"])
	require.NotEmpty(t, lines)
	spans := asSlice(t, lines[0])
	require.NotEmpty(t, spans)
	return asMap(t, spans[0])["text"].(string)
}

func firstKind(t *testing.T, res any) string {
	t.Helper()
	blocks := asSlice(t, asMap(t, res)["blocks"])
	require.NotEmpty(t, blocks)
	return asMap(t, blocks[0])["kind"].(string)
}

func TestResult(t *testing.T) {
	t.Run("maps every field with lowercase names", func(t *testing.T) {
		got := Result(shell.Result{
			Blocks: []shell.Block{{
				Kind: shell.BlockMedia,
				Src:  "/s",
				Alt:  "a",
				Lines: []shell.Line{
					{{Text: "x", Style: shell.StyleAccent, Href: "https://e.test"}, {Text: "y", Style: shell.StyleDim, Run: "ls"}},
					{},
				},
			}},
			Effects: []shell.Effect{{Kind: shell.EffectTheme, Arg: "amber"}},
		})

		want := map[string]any{
			"blocks": []any{map[string]any{
				"kind": "media", "src": "/s", "alt": "a",
				"lines": []any{
					[]any{
						map[string]any{"text": "x", "style": "accent", "href": "https://e.test", "run": ""},
						map[string]any{"text": "y", "style": "dim", "href": "", "run": "ls"},
					},
					[]any{},
				},
			}},
			"effects": []any{map[string]any{"kind": "theme", "arg": "amber"}},
		}
		assert.Equal(t, want, got)
	})

	t.Run("an empty result has empty arrays, never nil", func(t *testing.T) {
		got := Result(shell.Result{})

		assert.Equal(t, map[string]any{"blocks": []any{}, "effects": []any{}}, got)
	})

	t.Run("a block without lines has an empty lines array", func(t *testing.T) {
		got := Result(shell.Result{Blocks: []shell.Block{{Kind: shell.BlockError}}})

		block := asMap(t, asSlice(t, got["blocks"])[0])
		assert.Equal(t, []any{}, block["lines"])
		assert.Equal(t, "error", block["kind"])
	})

	t.Run("every style and effect kind serialises by name", func(t *testing.T) {
		styles := []string{"plain", "bold", "dim", "accent", "success", "warning", "error", "heading", "path"}
		var spans shell.Line
		for i := range styles {
			spans = append(spans, shell.Span{Text: "t", Style: shell.Style(i)})
		}
		effects := []shell.Effect{{Kind: shell.EffectClear}, {Kind: shell.EffectOpen}, {Kind: shell.EffectFetch}, {Kind: shell.EffectTheme}}

		got := Result(shell.Result{Blocks: []shell.Block{{Lines: []shell.Line{spans}}}, Effects: effects})

		line := asSlice(t, asSlice(t, asMap(t, asSlice(t, got["blocks"])[0])["lines"])[0])
		for i, name := range styles {
			assert.Equal(t, name, asMap(t, line[i])["style"])
		}
		for i, name := range []string{"clear", "open", "fetch", "theme"} {
			assert.Equal(t, name, asMap(t, asSlice(t, got["effects"])[i])["kind"])
		}
	})
}

func TestCompletion(t *testing.T) {
	t.Run("maps line, cursor and candidates", func(t *testing.T) {
		got := Completion(shell.Completion{Line: "ca", Cursor: 2, Candidates: []string{"cat", "camera"}})

		assert.Equal(t, map[string]any{"line": "ca", "cursor": 2, "candidates": []any{"cat", "camera"}}, got)
	})

	t.Run("no candidates is an empty array", func(t *testing.T) {
		got := Completion(shell.Completion{Line: "zz", Cursor: 2})

		assert.Equal(t, []any{}, got["candidates"])
	})
}

func TestCall(t *testing.T) {
	t.Run("boot returns the first screen", func(t *testing.T) {
		got := newBridge().Call("boot", nil)

		assert.NotEmpty(t, asSlice(t, asMap(t, got)["blocks"]))
	})

	t.Run("exec runs a command", func(t *testing.T) {
		got := newBridge().Call("exec", []any{"whoami"})

		assert.Equal(t, "visitor", firstText(t, got))
	})

	t.Run("exec records history", func(t *testing.T) {
		b := newBridge()
		b.Call("exec", []any{"pwd"})

		assert.Equal(t, []any{"pwd"}, b.Call("history", nil))
	})

	t.Run("history is an empty array before any command", func(t *testing.T) {
		assert.Equal(t, []any{}, newBridge().Call("history", nil))
	})

	t.Run("complete returns the completion shape", func(t *testing.T) {
		got := asMap(t, newBridge().Call("complete", []any{"ca", float64(2)}))

		assert.Contains(t, got, "line")
		assert.Contains(t, got, "cursor")
		assert.NotEmpty(t, asSlice(t, got["candidates"]))
	})

	t.Run("complete passes the cursor through", func(t *testing.T) {
		got := asMap(t, newBridge().Call("complete", []any{"ls zzz", float64(6)}))

		assert.Equal(t, 6, got["cursor"])
	})

	t.Run("prompt returns a string", func(t *testing.T) {
		assert.Equal(t, "visitor@pascalallen.com:~$", newBridge().Call("prompt", nil))
	})

	t.Run("commands lists name, synopsis and summary", func(t *testing.T) {
		got := asSlice(t, newBridge().Call("commands", nil))

		require.NotEmpty(t, got)
		first := asMap(t, got[0])
		assert.NotEmpty(t, first["name"])
		assert.Contains(t, first, "synopsis")
		assert.Contains(t, first, "summary")
	})

	t.Run("deliver renders feed items", func(t *testing.T) {
		item := map[string]any{"name": "pgqueue", "description": "queue", "url": "https://github.com/pascalallen/pgqueue", "meta": "Go"}

		got := newBridge().Call("deliver", []any{"repos", []any{item}})

		assert.Equal(t, "Repositories on GitHub", firstText(t, got))
	})

	t.Run("deliver passes every item field through", func(t *testing.T) {
		item := map[string]any{"name": "pgqueue", "description": "a queue", "url": "https://github.com/pascalallen/pgqueue", "meta": "Go"}

		got := newBridge().Call("deliver", []any{"repos", []any{item}})

		lines := asSlice(t, asMap(t, asSlice(t, asMap(t, got)["blocks"])[0])["lines"])
		link := asMap(t, asSlice(t, lines[1])[0])
		assert.Equal(t, "pgqueue", link["text"])
		assert.Equal(t, "https://github.com/pascalallen/pgqueue", link["href"])
		assert.Equal(t, "  Go", asMap(t, asSlice(t, lines[1])[1])["text"])
		assert.Equal(t, "  a queue", asMap(t, asSlice(t, lines[2])[0])["text"])
	})

	t.Run("deliver with no items is still a result", func(t *testing.T) {
		got := newBridge().Call("deliver", []any{"repos", []any{}})

		assert.Contains(t, asMap(t, got), "blocks")
	})

	t.Run("deliverError renders the failure", func(t *testing.T) {
		got := newBridge().Call("deliverError", []any{"repos"})

		assert.Equal(t, "error", firstKind(t, got))
		assert.Equal(t, "repos: could not load the list right now", firstText(t, got))
	})

	t.Run("an unknown feed is an error block from the shell", func(t *testing.T) {
		got := newBridge().Call("deliverError", []any{"nope"})

		assert.Equal(t, "error", firstKind(t, got))
	})
}

func TestCallRejectsBadArguments(t *testing.T) {
	good := map[string]any{"name": "n", "description": "d", "url": "u", "meta": "m"}
	cases := []struct {
		name string
		call string
		args []any
	}{
		{"exec with no arguments", "exec", nil},
		{"exec with two arguments", "exec", []any{"a", "b"}},
		{"exec with a number", "exec", []any{float64(1)}},
		{"exec with null", "exec", []any{nil}},
		{"complete with one argument", "complete", []any{"ca"}},
		{"complete with a string cursor", "complete", []any{"ca", "2"}},
		{"complete with a fractional cursor", "complete", []any{"ca", 1.5}},
		{"complete with a negative cursor", "complete", []any{"ca", float64(-1)}},
		{"complete with a huge cursor", "complete", []any{"ca", 1e30}},
		{"complete with a cursor one past the limit", "complete", []any{"ca", float64(2147483648)}},
		{"complete with a non-string line", "complete", []any{true, float64(1)}},
		{"boot with arguments", "boot", []any{"x"}},
		{"prompt with arguments", "prompt", []any{"x"}},
		{"history with arguments", "history", []any{"x"}},
		{"commands with arguments", "commands", []any{"x"}},
		{"deliver with one argument", "deliver", []any{"repos"}},
		{"deliver with a non-array", "deliver", []any{"repos", "x"}},
		{"deliver with a non-string feed", "deliver", []any{float64(1), []any{}}},
		{"deliver with a non-object item", "deliver", []any{"repos", []any{"x"}}},
		{"deliver with a missing field", "deliver", []any{"repos", []any{map[string]any{"name": "n"}}}},
		{"deliver with a non-string field", "deliver", []any{"repos", []any{map[string]any{"name": 1.0, "description": "d", "url": "u", "meta": "m"}}}},
		{"deliver with one bad item among good ones", "deliver", []any{"repos", []any{good, "x"}}},
		{"deliverError with no arguments", "deliverError", nil},
		{"deliverError with a number", "deliverError", []any{float64(1)}},
		{"an unknown function", "nope", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name+" returns an error result", func(t *testing.T) {
			got := asMap(t, newBridge().Call(tc.call, tc.args))

			blocks := asSlice(t, got["blocks"])
			require.Len(t, blocks, 1)
			assert.Equal(t, "error", asMap(t, blocks[0])["kind"])
			assert.Equal(t, []any{}, got["effects"])
			assert.Contains(t, firstText(t, got), "shell: ")
		})
	}

	t.Run("a rejected exec does not reach history", func(t *testing.T) {
		b := newBridge()
		b.Call("exec", []any{float64(1)})

		assert.Equal(t, []any{}, b.Call("history", nil))
	})

	t.Run("the largest valid cursor is accepted", func(t *testing.T) {
		got := asMap(t, newBridge().Call("complete", []any{"ls", float64(2147483647)}))

		assert.Contains(t, got, "candidates")
	})

	t.Run("a cursor of zero is accepted", func(t *testing.T) {
		got := asMap(t, newBridge().Call("complete", []any{"ls", float64(0)}))

		assert.Contains(t, got, "candidates")
	})
}

func TestCallRecoversFromPanics(t *testing.T) {
	t.Run("a panic inside the shell becomes an error result", func(t *testing.T) {
		b := &Bridge{}

		got := b.Call("exec", []any{"ls"})

		assert.Equal(t, "error", firstKind(t, got))
		assert.Equal(t, "shell: exec failed", firstText(t, got))
	})
}
