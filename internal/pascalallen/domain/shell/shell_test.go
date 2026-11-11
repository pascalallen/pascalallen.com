package shell

import (
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrompt(t *testing.T) {
	t.Run("shows the home directory as a tilde", func(t *testing.T) {
		assert.Equal(t, "visitor@pascalallen.com:~$", newTestShell().Prompt())
	})

	t.Run("shows the working directory below home", func(t *testing.T) {
		s := newTestShell()
		s.Exec("cd writing")
		assert.Equal(t, "visitor@pascalallen.com:~/writing$", s.Prompt())
	})

	t.Run("uses the injected user and host names", func(t *testing.T) {
		s := newTestShell(WithUser("guest"), WithHost("example.test"))
		assert.Equal(t, "guest@example.test:~$", s.Prompt())
		assert.Equal(t, "guest", resultText(s.Exec("whoami")))
		assert.Equal(t, "/home/guest", resultText(s.Exec("pwd")))
	})

	t.Run("ignores empty option values", func(t *testing.T) {
		s := New(fstest.MapFS{}, WithUser(""), WithHost(""), WithClock(nil), WithTheme("pink"))
		assert.Equal(t, "visitor@pascalallen.com:~$", s.Prompt())
		assert.Equal(t, "Thu Jan  1 00:00:00 UTC 1970", resultText(s.Exec("date")))
		assert.Contains(t, resultText(s.Exec("theme")), "* dark")
	})

	t.Run("starts with the injected theme", func(t *testing.T) {
		assert.Contains(t, resultText(newTestShell(WithTheme("amber")).Exec("theme")), "* amber")
	})

	t.Run("SetTheme switches to a known theme without output or history", func(t *testing.T) {
		s := newTestShell()

		assert.True(t, s.SetTheme("light"))

		assert.Contains(t, resultText(s.Exec("theme")), "* light")
		assert.Equal(t, []string{"theme"}, s.History())
	})

	t.Run("SetTheme ignores an unknown theme", func(t *testing.T) {
		s := newTestShell(WithTheme("amber"))

		assert.False(t, s.SetTheme("neon"))

		assert.Contains(t, resultText(s.Exec("theme")), "* amber")
	})

	t.Run("strips control characters from the prompt", func(t *testing.T) {
		assert.Equal(t, "a b@pascalallen.com:~$", newTestShell(WithUser("a\x1bb")).Prompt())
	})
}

func TestPathConfinement(t *testing.T) {
	t.Run("cd above home stays at home", func(t *testing.T) {
		s := newTestShell()
		s.Exec("cd writing")
		s.Exec("cd ../../..")
		assert.Equal(t, "/home/visitor", resultText(s.Exec("pwd")))
		assert.Equal(t, "visitor@pascalallen.com:~$", s.Prompt())
	})

	t.Run("cat above home resolves inside the tree", func(t *testing.T) {
		s := newTestShell()
		assert.Equal(t, "cat: ../../etc/passwd: No such file or directory", errorText(s.Exec("cat ../../etc/passwd")))
		assert.Contains(t, resultText(s.Exec("cat ../../about.md")), "Pascal Allen")
	})

	t.Run("a tilde path cannot climb out of home", func(t *testing.T) {
		s := newTestShell()
		assert.Equal(t, "cat: ~/../x: No such file or directory", errorText(s.Exec("cat ~/../x")))
		assert.Contains(t, resultText(s.Exec("cat ~/../about.md")), "Pascal Allen")
	})

	t.Run("an absolute path is rooted at home", func(t *testing.T) {
		s := newTestShell()
		assert.Contains(t, resultText(s.Exec("cat /about.md")), "Pascal Allen")
		assert.Equal(t, "cat: /etc/passwd: No such file or directory", errorText(s.Exec("cat /etc/passwd")))
		s.Exec("cd /")
		assert.Equal(t, "/home/visitor", resultText(s.Exec("pwd")))
	})

	t.Run("resolve never leaves the tree", func(t *testing.T) {
		s := newTestShell()
		s.cwd = "writing"
		for in, want := range map[string]string{
			"":                              "writing",
			".":                             "writing",
			"..":                            ".",
			"../..":                         ".",
			"../../../projects":             "projects",
			"~":                             ".",
			"~/..":                          ".",
			"/":                             ".",
			"/home/visitor/x":               "x",
			"/home/visitor":                 ".",
			"~/../../../projects/pgmcp.md":  "projects/pgmcp.md",
			"..//./projects/../experience/": "experience",
		} {
			got, problem := s.resolve(in)
			assert.Equal(t, want, got, in)
			assert.Empty(t, problem, in)
		}
	})

	t.Run("resolve refuses to descend through a file or a missing directory", func(t *testing.T) {
		s := newTestShell()
		for in, want := range map[string]string{
			"about.md/":        problemNotDir,
			"about.md/..":      problemNotDir,
			"~/about.md/x":     problemNotDir,
			"nope/":            problemMissing,
			"nope/../about.md": problemMissing,
			"/home/visitorx/a": problemMissing,
		} {
			_, problem := s.resolve(in)
			assert.Equal(t, want, problem, in)
		}
	})
}

func TestSequencing(t *testing.T) {
	t.Run("semicolons run every command", func(t *testing.T) {
		r := newTestShell().Exec("cat nope; echo after")
		assert.Contains(t, resultText(r), "after")
	})

	t.Run("and-and stops after a failure", func(t *testing.T) {
		r := newTestShell().Exec("cat nope && echo skipped && echo also")
		assert.NotContains(t, resultText(r), "skipped")
		assert.NotContains(t, resultText(r), "also")
	})

	t.Run("a semicolon after a skipped command runs again", func(t *testing.T) {
		r := newTestShell().Exec("cat nope && echo skipped; echo resumed")
		assert.NotContains(t, resultText(r), "skipped")
		assert.Contains(t, resultText(r), "resumed")
	})

	t.Run("clear drops output that came before it", func(t *testing.T) {
		r := newTestShell().Exec("echo before; clear; echo after")
		assert.Equal(t, "after", resultText(r))
		assert.Equal(t, []Effect{{Kind: EffectClear}}, r.Effects)
	})

	t.Run("blocks of the same kind merge and kinds alternate", func(t *testing.T) {
		r := newTestShell().Exec("echo a; echo b; cat nope; echo c")
		require.Len(t, r.Blocks, 3)
		assert.Equal(t, []BlockKind{BlockText, BlockError, BlockText}, []BlockKind{r.Blocks[0].Kind, r.Blocks[1].Kind, r.Blocks[2].Kind})
		assert.Len(t, r.Blocks[0].Lines, 2)
	})

	t.Run("empty input returns nothing", func(t *testing.T) {
		assert.Equal(t, Result{}, newTestShell().Exec("   "))
	})

	t.Run("unsupported syntax is one error line", func(t *testing.T) {
		r := newTestShell().Exec("cat about.md | grep Go")
		assert.Equal(t, "bash: pipes are not supported in this shell", errorText(r))
		assert.Len(t, r.Blocks, 1)
	})
}

func TestUnknownCommand(t *testing.T) {
	t.Run("reports command not found", func(t *testing.T) {
		r := newTestShell().Exec("frobnicate")
		assert.Equal(t, "bash: frobnicate: command not found", errorText(r))
	})

	t.Run("suggests the nearest command", func(t *testing.T) {
		r := newTestShell().Exec("sl")
		assert.Equal(t, "bash: sl: command not found\ndid you mean ls?", errorText(r))
		assert.Equal(t, []string{"ls"}, runs(r))
	})

	t.Run("suggests a command one typo away", func(t *testing.T) {
		assert.Contains(t, errorText(newTestShell().Exec("histroy")), "did you mean history?")
	})

	t.Run("does not suggest hidden commands", func(t *testing.T) {
		assert.NotContains(t, errorText(newTestShell().Exec("sudoo")), "did you mean")
	})

	t.Run("does not suggest for a one-letter word", func(t *testing.T) {
		assert.Equal(t, "bash: x: command not found", errorText(newTestShell().Exec("x")))
	})
}

func TestHistory(t *testing.T) {
	t.Run("records lines in order, trimmed", func(t *testing.T) {
		s := newTestShell()
		s.Exec("  ls ")
		s.Exec("pwd")
		assert.Equal(t, []string{"ls", "pwd"}, s.History())
	})

	t.Run("collapses consecutive duplicates", func(t *testing.T) {
		s := newTestShell()
		s.Exec("ls")
		s.Exec("ls")
		s.Exec("pwd")
		s.Exec("ls")
		assert.Equal(t, []string{"ls", "pwd", "ls"}, s.History())
	})

	t.Run("skips blank lines and keeps failing ones", func(t *testing.T) {
		s := newTestShell()
		s.Exec("")
		s.Exec("   ")
		s.Exec("nope")
		s.Exec(`echo "x`)
		assert.Equal(t, []string{"nope", `echo "x`}, s.History())
	})

	t.Run("keeps only the most recent entries", func(t *testing.T) {
		s := newTestShell()
		for i := 0; i < historyLimit+50; i++ {
			s.Exec("echo " + strconv.Itoa(i))
		}
		h := s.History()
		require.Len(t, h, historyLimit)
		assert.Equal(t, "echo 50", h[0])
		assert.Equal(t, "echo "+strconv.Itoa(historyLimit+49), h[len(h)-1])
	})

	t.Run("returns a copy", func(t *testing.T) {
		s := newTestShell()
		s.Exec("ls")
		s.History()[0] = "changed"
		assert.Equal(t, []string{"ls"}, s.History())
	})
}

func TestBoot(t *testing.T) {
	s := newTestShell()
	r := s.Boot()

	t.Run("shows the banner, the about text and tappable commands", func(t *testing.T) {
		text := resultText(r)
		assert.True(t, strings.HasPrefix(text, "pascalallen.com"))
		assert.Contains(t, text, "Full-stack software engineer")
		assert.Subset(t, runs(r), []string{"help", "about", "resume", "writing", "repos", "camera"})
	})

	t.Run("does not touch history", func(t *testing.T) {
		assert.Empty(t, s.History())
	})

	t.Run("still boots without an about file", func(t *testing.T) {
		r := New(fstest.MapFS{}).Boot()
		assert.Contains(t, runs(r), "help")
		lines := strings.Split(resultText(r), "\n")
		require.Len(t, lines, 4)
		assert.Equal(t, "", lines[1])
		assert.True(t, strings.HasPrefix(lines[2], "help  about"))
	})
}

func TestDeliver(t *testing.T) {
	s := newTestShell()

	t.Run("renders repositories as links with meta and description", func(t *testing.T) {
		r := s.Deliver(FeedRepos, []FeedItem{
			{Name: "pgqueue", Description: "Postgres-backed job queue", URL: "https://github.com/pascalallen/pgqueue", Meta: "Go"},
			{Name: "bare"},
		})
		assert.Equal(t, "Repositories on GitHub\npgqueue  Go\n  Postgres-backed job queue\nbare", resultText(r))
		assert.Equal(t, []string{"https://github.com/pascalallen/pgqueue"}, hrefs(r))
	})

	t.Run("renders packages", func(t *testing.T) {
		r := s.Deliver(FeedPackages, []FeedItem{{Name: "@pascalallen/react-form-components", URL: "https://www.npmjs.com/package/@pascalallen/react-form-components", Meta: "v3.1.0"}})
		assert.Contains(t, resultText(r), "Packages on npm")
		assert.Contains(t, resultText(r), "v3.1.0")
	})

	t.Run("does not link an unsafe url", func(t *testing.T) {
		r := s.Deliver(FeedRepos, []FeedItem{{Name: "evil", URL: "javascript:alert(1)"}})
		assert.Empty(t, hrefs(r))
	})

	t.Run("says so when a feed is empty", func(t *testing.T) {
		assert.Contains(t, resultText(s.Deliver(FeedPackages, nil)), "nothing to show yet")
	})

	t.Run("rejects an unknown feed", func(t *testing.T) {
		assert.Equal(t, "weather: unknown feed", errorText(s.Deliver("weather", nil)))
		assert.Equal(t, "weather: unknown feed", errorText(s.DeliverError("weather")))
	})

	t.Run("reports a failed fetch with a retry command", func(t *testing.T) {
		r := s.DeliverError(FeedRepos)
		assert.Equal(t, "repos: could not load the list right now\ntry again: repos", errorText(r))
		assert.Equal(t, []string{"repos"}, runs(r))
	})

	t.Run("strips control characters from feed text", func(t *testing.T) {
		r := s.Deliver(FeedRepos, []FeedItem{{Name: "a\x1b[31mb", Description: "x\ny"}})
		assert.Equal(t, "Repositories on GitHub\na [31mb\n  x y", resultText(r))
	})
}

func TestCamera(t *testing.T) {
	t.Run("emits a media block with the stream url", func(t *testing.T) {
		r := newTestShell().Exec("camera")
		require.Len(t, r.Blocks, 1)
		assert.Equal(t, BlockMedia, r.Blocks[0].Kind)
		assert.Equal(t, "/api/v1/camera/stream", r.Blocks[0].Src)
		assert.NotEmpty(t, r.Blocks[0].Alt)
	})
}

func TestRendering(t *testing.T) {
	t.Run("headings lose their marker and take the heading style", func(t *testing.T) {
		lines := renderText("# Title\nbody\n")
		assert.Equal(t, []Line{{{Text: "Title", Style: StyleHeading}}, {{Text: "body"}}}, lines)
	})

	t.Run("urls and emails become links and punctuation stays text", func(t *testing.T) {
		line := linkify("see (https://a.test/x), or mail me@b.test.", StylePlain)
		assert.Equal(t, Line{
			{Text: "see ("},
			{Text: "https://a.test/x", Style: StyleAccent, Href: "https://a.test/x"},
			{Text: "), or mail "},
			{Text: "me@b.test", Style: StyleAccent, Href: "mailto:me@b.test"},
			{Text: "."},
		}, line)
	})

	t.Run("words that only look like links stay plain", func(t *testing.T) {
		for _, w := range []string{"https://", "@pascalallen/pkg", "@b.co", "a@b", "a@b.", "x@y@z.com", "javascript:alert(1)", "ftp://x.test"} {
			assert.Equal(t, Line{{Text: w}}, linkify(w, StylePlain), w)
		}
	})

	t.Run("email detection rejects malformed addresses", func(t *testing.T) {
		assert.True(t, isEmail("me@b.test"))
		for _, w := range []string{"@b.co", "a@.co", "a@b.", "a@bco", "a!b@c.de", "a b@c.de", "a@b@c.de"} {
			assert.False(t, isEmail(w), w)
		}
	})

	t.Run("an empty file renders no lines", func(t *testing.T) {
		assert.Nil(t, renderText("\n"))
	})

	t.Run("an empty line renders an empty line", func(t *testing.T) {
		assert.Equal(t, []Line{{{Text: "a"}}, {}, {{Text: "b"}}}, renderText("a\n\nb"))
	})

	t.Run("summaries join wrapped headings and take the first body line", func(t *testing.T) {
		title, sub := summarize("# One\n# Two\n\n  first  \nsecond\n")
		assert.Equal(t, "One Two", title)
		assert.Equal(t, "first", sub)
		_, sub = summarize("# One\n\n# Two\nbody\n")
		assert.Equal(t, "body", sub)
		title, sub = summarize("no heading\n")
		assert.Equal(t, "", title)
		assert.Equal(t, "no heading", sub)
	})

	t.Run("style and kind names are stable", func(t *testing.T) {
		assert.Equal(t, "heading", StyleHeading.String())
		assert.Equal(t, "path", StylePath.String())
		assert.Equal(t, "unknown", Style(99).String())
		assert.Equal(t, "media", BlockMedia.String())
		assert.Equal(t, "unknown", BlockKind(-1).String())
		assert.Equal(t, "fetch", EffectFetch.String())
		assert.Equal(t, "unknown", EffectKind(42).String())
	})
}

func TestMissingContent(t *testing.T) {
	s := New(fstest.MapFS{"odd dir/it's.md": {Data: []byte("# Odd\n")}})

	t.Run("shortcuts report a missing file", func(t *testing.T) {
		assert.Equal(t, "about: ~/about.md: No such file or directory", errorText(s.Exec("about")))
		assert.Equal(t, "writing: ~/writing: No such file or directory", errorText(s.Exec("writing")))
	})

	t.Run("run commands quote names with special characters", func(t *testing.T) {
		assert.Contains(t, runs(s.Exec("ls")), "ls '~/odd dir'")
		assert.Contains(t, runs(s.Exec("ls 'odd dir'")), `cat '~/odd dir/it'\''s.md'`)
	})

	t.Run("a quoted run command resolves", func(t *testing.T) {
		assert.Equal(t, "Odd", resultText(s.Exec(`cat '~/odd dir/it'\''s.md'`)))
	})

	t.Run("listing falls back to the file name without a heading", func(t *testing.T) {
		s := New(fstest.MapFS{"writing/plain.md": {Data: []byte("text\n")}, "writing/sub/x.md": {Data: []byte("x")}})
		r := s.Exec("writing")
		assert.Contains(t, resultText(r), "  plain.md")
		assert.NotContains(t, resultText(r), "sub")
	})
}

func TestSanitising(t *testing.T) {
	t.Run("echoed control characters become spaces", func(t *testing.T) {
		r := newTestShell().Exec("echo 'a\x07b\x1b[2Jc'")
		assert.Equal(t, "a b [2Jc", resultText(r))
	})

	t.Run("invalid utf-8 is replaced", func(t *testing.T) {
		r := newTestShell().Exec("echo \xff")
		assert.Equal(t, "�", resultText(r))
	})

	t.Run("unknown command names are sanitised", func(t *testing.T) {
		r := newTestShell().Exec("'\x00x'")
		assert.Equal(t, "bash:  x: command not found", errorText(r))
	})
}

func assertClean(t *testing.T, r Result) {
	t.Helper()
	check := func(s string) {
		if !utf8.ValidString(s) {
			t.Fatalf("invalid utf-8 in %q", s)
		}
		for _, c := range s {
			if c != 0x200C && c != 0x200D && unicode.In(c, unicode.Cc, unicode.Cf, unicode.Zl, unicode.Zp) {
				t.Fatalf("unsafe character %U in %q", c, s)
			}
		}
	}
	for _, b := range r.Blocks {
		check(b.Src)
		check(b.Alt)
		for _, l := range b.Lines {
			for _, sp := range l {
				check(sp.Text)
				check(sp.Href)
				check(sp.Run)
			}
		}
	}
	for _, e := range r.Effects {
		check(e.Arg)
	}
}

func seedLines() []string {
	var lines []string
	for _, tc := range parseCases {
		lines = append(lines, tc.input)
	}
	for _, cases := range commandCases {
		for _, tc := range cases {
			lines = append(lines, tc.setup...)
			lines = append(lines, tc.line)
		}
	}
	return lines
}

func FuzzExec(f *testing.F) {
	for _, l := range seedLines() {
		f.Add(l)
	}
	for _, l := range unsafeSeeds {
		f.Add("echo " + l)
	}
	f.Fuzz(func(t *testing.T, line string) {
		s := newTestShell()
		assertClean(t, s.Exec(line))
		assertClean(t, s.Exec("history"))
		assertClean(t, s.Exec("cat "+line))
		assertClean(t, s.Deliver(FeedRepos, []FeedItem{{Name: line, Description: line, URL: line, Meta: line}}))
		p := s.Prompt()
		assert.True(t, strings.HasPrefix(p, "visitor@pascalallen.com:~"))
		assert.True(t, strings.HasSuffix(p, "$"))
	})
}

func FuzzComplete(f *testing.F) {
	for _, tc := range completeCases {
		f.Add(tc.line, tc.cursor)
	}
	for _, l := range seedLines() {
		f.Add(l, len(l))
	}
	f.Fuzz(func(t *testing.T, line string, cursor int) {
		c := newTestShell().Complete(line, cursor)
		for _, cand := range c.Candidates {
			assertClean(t, Result{Blocks: []Block{{Lines: []Line{{{Text: cand}}}}}})
		}
		n := utf8.RuneCountInString(c.Line)
		if c.Cursor < 0 || c.Cursor > n {
			t.Fatalf("cursor %d outside line of %d runes", c.Cursor, n)
		}
	})
}
