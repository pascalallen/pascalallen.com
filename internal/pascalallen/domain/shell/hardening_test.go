package shell

import (
	"errors"
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var unsafeSeeds = []string{
	"a\u202Eb",
	"a\u202Ab\u202Bc\u202Cd\u202De",
	"a\u2066b\u2067c\u2068d\u2069e",
	"a\u200Bb\u200Cc\u200Dd\u200Ee\u200Ff",
	"\uFEFFbom",
	"line\u2028sep\u2029para",
	"soft\u00ADhyphen",
	"bell\x07esc\x1b[2J",
	"bad\xffutf8",
}

func TestSanitiserRule(t *testing.T) {
	t.Run("format characters are removed and separators become spaces", func(t *testing.T) {
		assert.Equal(t, "ab", clean("a\u202Eb"))
		assert.Equal(t, "abcde", clean("a\u2066b\u2067c\u2068d\u2069e"))
		assert.Equal(t, "abcdef", clean("a\u200Bb\u200Cc\u200Dd\u200Ee\u200Ff"))
		assert.Equal(t, "bom", clean("\uFEFFbom"))
		assert.Equal(t, "line sep para", clean("line\u2028sep\u2029para"))
		assert.Equal(t, "bell esc [2J", clean("bell\x07esc\x1b[2J"))
		assert.Equal(t, "bad�utf8", clean("bad\xffutf8"))
		assert.Equal(t, "Pascal · Go — Austin – TX ’", clean("Pascal · Go — Austin – TX ’"))
	})

	t.Run("echoed input is sanitised", func(t *testing.T) {
		for _, seed := range unsafeSeeds {
			r := newTestShell().Exec("echo '" + seed + "'")
			assertClean(t, r)
		}
		assert.Equal(t, "evil.exe txt", resultText(newTestShell().Exec("echo 'evil\u202E.exe\u2028txt'")))
	})

	t.Run("error messages quoting input are sanitised", func(t *testing.T) {
		r := newTestShell().Exec("cat 'a\u202Eb'")
		assert.Equal(t, "cat: ab: No such file or directory", errorText(r))
	})

	t.Run("file content is sanitised", func(t *testing.T) {
		s := New(fstest.MapFS{"x.md": {Data: []byte("safe\u202Etext\u200B\n")}})
		assert.Equal(t, "safetext", resultText(s.Exec("cat x.md")))
	})

	t.Run("feed items are sanitised", func(t *testing.T) {
		r := newTestShell().Deliver(FeedRepos, []FeedItem{{Name: "re\u202Epo", Description: "d\u2029e", Meta: "\uFEFFGo", URL: "https://x.test/\u200Ba"}})
		assertClean(t, r)
		assert.Equal(t, []string{"https://x.test/a"}, hrefs(r))
	})

	t.Run("completion candidates are sanitised", func(t *testing.T) {
		s := New(fstest.MapFS{"a\u202Eone.md": {}, "a\u200Btwo.md": {}})
		assert.Equal(t, []string{"aone.md", "atwo.md"}, s.Complete("cat a", 5).Candidates)
	})
}

func TestLimits(t *testing.T) {
	t.Run("an over-long line is one error block and is not recorded", func(t *testing.T) {
		s := newTestShell()
		r := s.Exec("echo " + strings.Repeat("x", 1_000_000))
		require.Len(t, r.Blocks, 1)
		assert.Equal(t, "bash: input line too long (max "+strconv.Itoa(maxLineRunes)+" characters)", errorText(r))
		assert.Empty(t, s.History())
	})

	t.Run("a line at the limit still runs", func(t *testing.T) {
		line := "echo " + strings.Repeat("x", maxLineRunes-5)
		assert.Empty(t, errorText(newTestShell().Exec(line)))
	})

	t.Run("too many sequenced commands is one error block", func(t *testing.T) {
		r := newTestShell().Exec(strings.Repeat("tree;", 20000))
		require.Len(t, r.Blocks, 1)
		assert.Contains(t, errorText(r), "bash: input line too long")
		r = newTestShell().Exec(strings.Repeat("pwd;", maxSteps+1))
		require.Len(t, r.Blocks, 1)
		assert.Equal(t, "bash: too many commands on one line (max "+strconv.Itoa(maxSteps)+")", errorText(r))
	})

	t.Run("the maximum number of commands still runs", func(t *testing.T) {
		r := newTestShell().Exec(strings.Repeat("pwd;", maxSteps))
		assert.Len(t, strings.Split(resultText(r), "\n"), maxSteps)
	})

	t.Run("output is capped with a truncation notice", func(t *testing.T) {
		r := newTestShell().Exec(strings.Repeat("cat resume.md;", maxSteps))
		lines := strings.Split(resultText(r), "\n")
		require.Len(t, lines, maxOutputLines+1)
		assert.Equal(t, "output truncated after "+strconv.Itoa(maxOutputLines)+" lines", lines[len(lines)-1])
	})

	t.Run("deliver output is capped too", func(t *testing.T) {
		items := make([]FeedItem, maxOutputLines)
		for i := range items {
			items[i] = FeedItem{Name: "n", Description: "d"}
		}
		lines := strings.Split(resultText(newTestShell().Deliver(FeedRepos, items)), "\n")
		assert.Len(t, lines, maxOutputLines+1)
	})

	t.Run("a long unknown word gets no suggestion", func(t *testing.T) {
		word := strings.Repeat("l", maxSuggestRunes+1)
		assert.Equal(t, "bash: "+word+": command not found", errorText(newTestShell().Exec(word)))
	})

	t.Run("suggestions skip commands whose length differs by more than two", func(t *testing.T) {
		assert.Equal(t, "", newTestShell().nearest("expxxxxxxxxxxx"))
		assert.Equal(t, "experience", newTestShell().nearest("experiense"))
	})

	t.Run("completing an over-long line leaves it alone", func(t *testing.T) {
		line := "echo " + strings.Repeat("a", maxLineRunes) + "; whoa"
		c := newTestShell().Complete(line, len(line))
		assert.Equal(t, Completion{Line: line, Cursor: len(line)}, c)
	})
}

type brokenFS struct {
	fstest.MapFS
}

func (brokenFS) ReadFile(string) ([]byte, error) { return nil, errors.New("broken") }

func (brokenFS) ReadDir(string) ([]fs.DirEntry, error) { return nil, errors.New("broken") }

func TestUnreadableFiles(t *testing.T) {
	s := New(brokenFS{fstest.MapFS{"about.md": {Data: []byte("x")}, "writing/a.md": {Data: []byte("# A")}}})

	t.Run("a file that stats but cannot be read is reported", func(t *testing.T) {
		assert.Equal(t, "cat: about.md: No such file or directory", errorText(s.Exec("cat about.md")))
	})

	t.Run("a directory that cannot be listed is reported", func(t *testing.T) {
		assert.Equal(t, "ls: cannot open directory 'writing'", errorText(s.Exec("ls writing")))
	})

	t.Run("a directory that cannot be listed completes to nothing", func(t *testing.T) {
		assert.Equal(t, Completion{Line: "cat writing/", Cursor: 12}, s.Complete("cat writing/", 12))
	})

	t.Run("an unreadable tree counts nothing", func(t *testing.T) {
		assert.Contains(t, resultText(s.Exec("tree")), "0 directories, 0 files")
	})

	t.Run("shortcuts report unreadable content", func(t *testing.T) {
		assert.Equal(t, "about: ~/about.md: No such file or directory", errorText(s.Exec("about")))
		assert.Equal(t, "writing: ~/writing: No such file or directory", errorText(s.Exec("writing")))
	})

	t.Run("grep skips unreadable files", func(t *testing.T) {
		r := s.Exec("grep ''")
		assert.Empty(t, resultText(r))
	})

	t.Run("ls -l skips entries whose info cannot be read", func(t *testing.T) {
		s := New(badInfoFS{fstest.MapFS{"a.md": {Data: []byte("a")}}})
		assert.Equal(t, "-r--r--r--  visitor       1  a.md", resultText(s.Exec("ls -l")))
	})
}

type badInfoFS struct {
	fstest.MapFS
}

type badEntry struct{}

func (badEntry) Name() string               { return "bad.md" }
func (badEntry) IsDir() bool                { return false }
func (badEntry) Type() fs.FileMode          { return 0 }
func (badEntry) Info() (fs.FileInfo, error) { return nil, errors.New("broken") }

func (b badInfoFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := b.MapFS.ReadDir(name)
	return append(entries, badEntry{}), err
}

func TestListingDetail(t *testing.T) {
	t.Run("head separates several files with one blank line", func(t *testing.T) {
		r := newTestShell().Exec("head -n 1 about.md contact.md")
		assert.Equal(t, "==> about.md <==\nPascal Allen\n\n==> contact.md <==\nContact", resultText(r))
	})

	t.Run("an experience entry without a body line has no detail line", func(t *testing.T) {
		s := New(fstest.MapFS{"experience/a.md": {Data: []byte("# A\n")}})
		lines := strings.Split(resultText(s.Exec("experience")), "\n")
		assert.Equal(t, []string{"Experience", "  A", ""}, lines[:3])
	})

	t.Run("edit distance counts a transposition once and nothing cheaper", func(t *testing.T) {
		assert.Equal(t, 1, distance("sl", "ls"))
		assert.Equal(t, 3, distance("lop", "help"))
		assert.Equal(t, 3, distance("rhl", "help"))
		assert.Equal(t, 3, distance("ca", "abc"))
	})
}

func TestHistoryNumbering(t *testing.T) {
	t.Run("entry numbers stay absolute after the bound trims old entries", func(t *testing.T) {
		s := newTestShell()
		for i := 1; i <= historyLimit+50; i++ {
			s.Exec("echo " + strconv.Itoa(i))
		}
		r := s.Exec("history 2")
		assert.Equal(t, "  250  echo 250\n  251  history 2", resultText(r))
	})

	t.Run("collapsed duplicates do not advance the numbering", func(t *testing.T) {
		s := newTestShell()
		s.Exec("ls")
		s.Exec("ls")
		assert.Equal(t, "    1  ls\n    2  history", resultText(s.Exec("history")))
	})
}

func TestCaseFoldingHighlight(t *testing.T) {
	t.Run("a line whose lower case changes length is printed whole", func(t *testing.T) {
		s := New(fstest.MapFS{"x.md": {Data: []byte("İstanbul ok\n")}})
		r := s.Exec("grep -i ok x.md")
		assert.Equal(t, "x.md:1: İstanbul ok", resultText(r))
	})

	t.Run("a line whose lower case changes length gets no highlight", func(t *testing.T) {
		s := New(fstest.MapFS{"x.md": {Data: []byte("İstanbul ok\n")}})
		for _, sp := range spans(s.Exec("grep -i ok x.md")) {
			assert.NotEqual(t, StyleWarning, sp.Style, sp.Text)
		}
	})

	t.Run("highlight splits a line into exact spans", func(t *testing.T) {
		assert.Equal(t, []Span{styled("Go", StyleWarning), plain(" and "), styled("GO", StyleWarning)}, highlight("Go and GO", "go and go", "go"))
		assert.Equal(t, []Span{plain("a "), styled("b", StyleWarning), plain(" c")}, highlight("a b c", "a b c", "b"))
	})

	t.Run("an empty pattern matches every line without highlighting", func(t *testing.T) {
		r := newTestShell().Exec("grep '' contact.md")
		assert.Len(t, strings.Split(resultText(r), "\n"), 7)
		assert.Equal(t, []Span{plain("x")}, highlight("x", "x", ""))
	})

	t.Run("a matching line is highlighted when lengths agree", func(t *testing.T) {
		s := New(fstest.MapFS{"x.md": {Data: []byte("Go and GO\n")}})
		r := s.Exec("grep -i go x.md")
		var marked []string
		for _, sp := range spans(r) {
			if sp.Style == StyleWarning {
				marked = append(marked, sp.Text)
			}
		}
		assert.Equal(t, []string{"Go", "GO"}, marked)
	})
}
