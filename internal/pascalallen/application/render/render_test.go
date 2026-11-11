package render

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"html"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

var update = flag.Bool("update", false, "rewrite the golden boot html in testdata")

func render(blocks ...shell.Block) string {
	return string(Blocks(blocks))
}

func textBlock(spans ...shell.Span) shell.Block {
	return shell.Block{Kind: shell.BlockText, Lines: []shell.Line{spans}}
}

func TestSpans(t *testing.T) {
	t.Run("plain text is escaped", func(t *testing.T) {
		out := render(textBlock(shell.Span{Text: `<script>alert("x")</script> & 'y'`}))

		assert.NotContains(t, out, "<script>")
		assert.Contains(t, out, "&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt; &amp; &#39;y&#39;")
	})

	t.Run("style becomes a class and unknown styles fall back to plain", func(t *testing.T) {
		out := render(textBlock(
			shell.Span{Text: "a", Style: shell.StyleAccent},
			shell.Span{Text: "b", Style: shell.Style(99)},
		))

		assert.Contains(t, out, `<span class="s-accent">a</span>`)
		assert.Contains(t, out, `<span class="s-plain">b</span>`)
	})

	t.Run("a span of two or more spaces gets the gap class", func(t *testing.T) {
		out := render(textBlock(shell.Span{Text: "  "}, shell.Span{Text: " "}))

		assert.Contains(t, out, `<span class="s-plain gap">  </span>`)
		assert.Contains(t, out, `<span class="s-plain"> </span>`)
	})

	t.Run("allowed links carry target and rel and a normalised href", func(t *testing.T) {
		out := render(textBlock(
			shell.Span{Text: "a", Href: "https://example.com"},
			shell.Span{Text: "b", Href: "http://example.com/x?q=1&r=2"},
			shell.Span{Text: "c", Href: "mailto:me@example.com"},
		))

		assert.Contains(t, out, `<a class="s-plain" href="https://example.com/" target="_blank" rel="noopener noreferrer">a</a>`)
		assert.Contains(t, out, `href="http://example.com/x?q=1&amp;r=2"`)
		assert.Contains(t, out, `<a class="s-plain" href="mailto:me@example.com" target="_blank" rel="noopener noreferrer">c</a>`)
	})

	t.Run("refused schemes keep the text and drop the link", func(t *testing.T) {
		for _, href := range []string{"javascript:alert(1)", "data:text/html,x", "ftp://example.com", "https:", "http://[::1", "mailto:", "//example.com", "/relative", "JaVaScRiPt:alert(1)"} {
			out := render(textBlock(shell.Span{Text: "label", Href: href}))

			assert.NotContains(t, out, "<a ", href)
			assert.NotContains(t, out, "href", href)
			assert.Contains(t, out, `<span class="s-plain">label</span>`, href)
		}
	})

	t.Run("attribute values are escaped", func(t *testing.T) {
		out := render(textBlock(
			shell.Span{Text: "l", Href: `https://example.com/"onmouseover="x`},
			shell.Span{Text: "r", Run: `cat "a"><b>`},
		))

		assert.NotContains(t, out, `"onmouseover="x`)
		assert.NotContains(t, out, "<b>")
		assert.Contains(t, out, `data-run="cat &#34;a&#34;&gt;&lt;b&gt;"`)
	})

	t.Run("run spans are deep link anchors with the command encoded", func(t *testing.T) {
		out := render(textBlock(shell.Span{Text: "adk", Style: shell.StyleAccent, Run: "cat ~/projects/adk agent.md"}))

		assert.Contains(t, out, `<a class="run s-accent" href="/#cat%20~%2Fprojects%2Fadk%20agent.md" data-run="cat ~/projects/adk agent.md">adk</a>`)
	})
}

func TestBlocks(t *testing.T) {
	t.Run("kinds map to block classes and unknown kinds are text", func(t *testing.T) {
		out := render(
			shell.Block{Kind: shell.BlockError, Lines: []shell.Line{{{Text: "e"}}}},
			shell.Block{Kind: shell.BlockKind(9), Lines: []shell.Line{{{Text: "u"}}}},
		)

		assert.Contains(t, out, `<div class="block block-error"><div class="line"><span class="s-plain">e</span></div></div>`)
		assert.Contains(t, out, `<div class="block block-text"><div class="line"><span class="s-plain">u</span></div></div>`)
	})

	t.Run("an empty line is an empty line element", func(t *testing.T) {
		out := render(shell.Block{Kind: shell.BlockText, Lines: []shell.Line{{}}})

		assert.Equal(t, `<div class="block block-text"><div class="line"></div></div>`, out)
	})

	t.Run("media renders an image only for a same-origin path", func(t *testing.T) {
		ok := render(shell.Block{Kind: shell.BlockMedia, Src: "/api/v1/camera/stream?x=1", Alt: `a "cam"`})
		assert.Contains(t, ok, `<img class="media" src="/api/v1/camera/stream?x=1" alt="a &#34;cam&#34;">`)

		for _, src := range []string{"https://evil.example/x.png", "//evil.example/x.png", "///evil.example/x", "/\\evil.example/x", "\\\\evil.example/x", "\\evil", "\t/ok.png", "\n/ok.png", " /ok.png", "/a\tb.png", "/a\nb.png", "/a\x7fb.png", "/%zz", "/", "", "javascript:alert(1)", "relative.png"} {
			out := render(shell.Block{Kind: shell.BlockMedia, Src: src, Lines: []shell.Line{{{Text: "caption"}}}})
			assert.NotContains(t, out, "<img", src)
			assert.Contains(t, out, "caption", src)
		}
	})
}

func TestBootEntry(t *testing.T) {
	t.Run("the boot entry is marked prerendered", func(t *testing.T) {
		out := string(BootEntry(shell.Result{Blocks: []shell.Block{textBlock(shell.Span{Text: "hi"})}}))

		assert.True(t, strings.HasPrefix(out, `<div class="entry entry-boot" data-prerendered="">`))
		assert.True(t, strings.HasSuffix(out, `</div>`))
	})

	t.Run("the boot screen matches the golden file", func(t *testing.T) {
		out := string(BootEntry(shell.New(content.FS()).Boot())) + "\n"
		const golden = "testdata/boot.golden.html"
		if *update {
			require.NoError(t, os.MkdirAll("testdata", 0o755))
			require.NoError(t, os.WriteFile(golden, []byte(out), 0o644))
		}
		want, err := os.ReadFile(golden)
		require.NoError(t, err)

		assert.Equal(t, string(want), out)
	})
}

func hostileFS() fstest.MapFS {
	return fstest.MapFS{
		"about.md":     {Data: []byte("# Pascal Allen\n\nEngineer · Go\n\nAbout <script>alert(1)</script> \"q\"\n")},
		`we"ird<b>.md`: {Data: []byte("# <img src=x onerror=alert(1)>\nbody \"quoted\" <script>x()</script> https://example.com/?a=1&b=2\n")},
		"dir/it's.md":  {Data: []byte("it's <svg/onload=alert(1)>\n")},
	}
}

func TestPlain(t *testing.T) {
	t.Run("every file in the content tree appears with its text", func(t *testing.T) {
		fsys := content.FS()
		out := string(Plain(fsys))
		text := html.UnescapeString(tags.ReplaceAllString(out, ""))

		var count int
		err := walkFiles(fsys, func(name string, data []byte) {
			count++
			assert.Contains(t, out, ">"+name+"<", name)
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimPrefix(line, "# ")
				if line != "" {
					assert.Contains(t, text, line, name)
				}
			}
		})
		require.NoError(t, err)
		assert.Greater(t, count, 30)
	})

	t.Run("every link in the content tree is a real anchor", func(t *testing.T) {
		out := string(Plain(content.FS()))

		assert.GreaterOrEqual(t, strings.Count(out, "medium.com"), 19)
		assert.Contains(t, out, `href="https://github.com/pascalallen"`)
	})

	t.Run("sections follow the portfolio order", func(t *testing.T) {
		out := string(Plain(content.FS()))
		last := -1
		for _, name := range []string{">about.md<", ">skills.md<", ">experience/", ">projects/", ">writing/", ">education.md<", ">contact.md<"} {
			i := strings.Index(out, name)
			require.GreaterOrEqual(t, i, 0, name)
			assert.Greater(t, i, last, name)
			last = i
		}
	})

	t.Run("hostile file names and bodies come out inert", func(t *testing.T) {
		out := string(Plain(hostileFS()))

		for _, raw := range []string{"<script", "<img", "<svg", "<b>", `onerror=alert(1)>`} {
			assert.NotContains(t, out, raw)
		}
		assert.Contains(t, out, "we&#34;ird&lt;b&gt;.md")
		assert.Contains(t, out, "dir/it&#39;s.md")
		assert.Contains(t, out, "&lt;script&gt;x()&lt;/script&gt;")
		assert.Contains(t, out, `href="https://example.com/?a=1&amp;b=2"`)
	})
}

var tags = regexp.MustCompile(`<[^>]*>`)

func walkFiles(fsys fs.FS, fn func(string, []byte)) error {
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		fn(p, data)
		return nil
	})
}

func decodeEnv(encoded string) (map[string]any, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	return out, json.Unmarshal(raw, &out)
}

func TestDescription(t *testing.T) {
	long := strings.Repeat("alpha beta gamma delta ", 12)
	fsys := fstest.MapFS{"about.md": {Data: []byte("# Name\n\nTitle · Go\n\n" + long + "\nsecond line of the first paragraph\n\nSECOND-PARAGRAPH\n")}}

	t.Run("it is the first paragraph cut on a word boundary within 160 characters", func(t *testing.T) {
		d := NewPage(fsys, Config{}).Index("/").Head.Description

		assert.NotContains(t, d, "SECOND-PARAGRAPH")
		assert.LessOrEqual(t, len([]rune(d)), 160)
		assert.True(t, strings.HasSuffix(d, "…"))
		assert.True(t, strings.HasSuffix(strings.TrimSuffix(d, "…"), "alpha") || strings.HasSuffix(strings.TrimSuffix(d, "…"), "beta") || strings.HasSuffix(strings.TrimSuffix(d, "…"), "gamma") || strings.HasSuffix(strings.TrimSuffix(d, "…"), "delta"))
	})

	t.Run("a single very long word is cut within the limit including the ellipsis", func(t *testing.T) {
		one := fstest.MapFS{"about.md": {Data: []byte("# N\n\nT · Go\n\n" + strings.Repeat("x", 400) + "\n")}}

		assert.Equal(t, 160, len([]rune(NewPage(one, Config{}).Index("/").Head.Description)))
	})

	t.Run("a short first paragraph is kept whole and later paragraphs are ignored", func(t *testing.T) {
		short := fstest.MapFS{"about.md": {Data: []byte("# N\n\nT · Go\n\nOne line\nand two.\n\nLater.\n")}}

		assert.Equal(t, "One line and two.", NewPage(short, Config{}).Index("/").Head.Description)
	})
}

func TestPage(t *testing.T) {
	cfg := Config{BaseURL: "https://example.com/", Env: "test", AssetHashes: map[string]string{"wasm": "abc123", "empty": ""}}

	t.Run("the index head is indexable with a canonical url", func(t *testing.T) {
		d := NewPage(content.FS(), cfg).Index("/")

		assert.Equal(t, "https://example.com/", d.Head.Canonical)
		assert.Empty(t, d.Head.Robots)
		assert.Equal(t, "abc123", d.WasmHash)
		assert.Equal(t, "visitor@pascalallen.com:~$", d.Prompt)
		assert.Contains(t, string(d.Boot), `data-prerendered=""`)
	})

	t.Run("any other path is noindex and has no canonical", func(t *testing.T) {
		for _, p := range []string{"/nope", "/api/v1/zzz", "/index.html"} {
			d := NewPage(content.FS(), cfg).Index(p)

			assert.Equal(t, "noindex", d.Head.Robots, p)
			assert.Empty(t, d.Head.Canonical, p)
		}
	})

	t.Run("the plain view is indexable with its own canonical", func(t *testing.T) {
		d := NewPage(content.FS(), cfg).PlainView()

		assert.Equal(t, "https://example.com/plain", d.Head.Canonical)
		assert.Empty(t, d.Head.Robots)
	})

	t.Run("the page config exposes only the base url, env and asset hashes", func(t *testing.T) {
		d := NewPage(content.FS(), cfg).Index("/")
		raw, err := decodeEnv(d.Environment)
		require.NoError(t, err)

		assert.Equal(t, map[string]any{
			"APP_BASE_URL": "https://example.com/",
			"APP_ENV":      "test",
			"ASSET_HASHES": map[string]any{"wasm": "abc123"},
		}, raw)
	})

	t.Run("json-ld is a person with profile links and skills", func(t *testing.T) {
		d := NewPage(content.FS(), cfg).Index("/")
		var doc map[string]any
		require.NoError(t, json.Unmarshal([]byte(d.Head.JSONLD), &doc))

		assert.Equal(t, "https://schema.org", doc["@context"])
		assert.Equal(t, "Person", doc["@type"])
		assert.Equal(t, "Pascal Allen", doc["name"])
		assert.Equal(t, "Senior Software Engineer", doc["jobTitle"])
		assert.Equal(t, "https://example.com", doc["url"])
		assert.ElementsMatch(t, []any{
			"https://github.com/pascalallen",
			"https://pascalallen.medium.com",
			"https://www.linkedin.com/in/pascal-allen-942749112",
		}, doc["sameAs"])
		assert.Contains(t, doc["knowsAbout"], "Go")
		assert.Contains(t, doc["knowsAbout"], "AES-GCM secret encryption")
	})

	t.Run("json-ld built from hostile content cannot close the script element", func(t *testing.T) {
		fsys := fstest.MapFS{
			"about.md":   {Data: []byte("# </script><script>alert(1)</script>\n\nEng</script> · Go\n\nprose\n")},
			"contact.md": {Data: []byte("GitHub javascript:alert(1)\nLinkedIn mailto:a@b.example\nMedium https://m.example/</script>\n")},
		}
		d := NewPage(fsys, cfg).Index("/")

		assert.NotContains(t, string(d.Head.JSONLD), "</script>")
		assert.NotContains(t, string(d.Head.JSONLD), "javascript:")
		assert.NotContains(t, string(d.Head.JSONLD), "mailto:")
		var doc map[string]any
		require.NoError(t, json.Unmarshal([]byte(d.Head.JSONLD), &doc))
	})

	t.Run("an empty base url omits canonical, url and sitemap entries", func(t *testing.T) {
		p := NewPage(content.FS(), Config{})
		d := p.Index("/")

		assert.Empty(t, d.Head.Canonical)
		assert.NotContains(t, string(d.Head.JSONLD), `"url"`)
		assert.NotContains(t, p.Robots(), "Sitemap")
		body, err := p.Sitemap()
		require.NoError(t, err)
		assert.NotContains(t, string(body), "<loc>")
	})

	t.Run("robots allows all and links the sitemap", func(t *testing.T) {
		assert.Equal(t, "User-agent: *\nAllow: /\n\nSitemap: https://example.com/sitemap.xml\n", NewPage(content.FS(), cfg).Robots())
	})

	t.Run("the description is the about prose and short", func(t *testing.T) {
		d := NewPage(content.FS(), cfg).Index("/")

		assert.True(t, strings.HasPrefix(d.Head.Description, "Full-stack software engineer"))
		assert.LessOrEqual(t, len([]rune(d.Head.Description)), 161)
	})
}
