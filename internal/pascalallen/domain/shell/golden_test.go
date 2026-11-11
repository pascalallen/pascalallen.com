package shell

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite the golden transcripts in testdata")

var session = []string{
	"help",
	"ls",
	"cd writing",
	"ls -l",
	"cat 03-live-security-camera-with-go.md",
	"cd ..",
	"grep -i raspberry projects",
	"experience",
	"cat ~/experience/01-sectigo.md",
	"head -n 4 resume.md",
	"tree projects",
	"man grep",
	"theme",
	"theme green",
	"repos",
	"packages",
	"camera",
	"open contact.md",
	"echo 'quoted  text' \"and more\" # a comment",
	"cd nope && echo skipped; echo next",
	"sl",
	"cat about.md | grep Go",
	`echo "unterminated`,
	"history 5",
	"rm -rf /",
	"sudo make me a sandwich",
}

func renderSpan(sp Span) string {
	switch {
	case sp.Href != "":
		return "[" + sp.Style.String() + ":" + sp.Text + " -> " + sp.Href + "]"
	case sp.Run != "":
		return "[" + sp.Style.String() + ":" + sp.Text + " => " + sp.Run + "]"
	case sp.Style != StylePlain:
		return "[" + sp.Style.String() + ":" + sp.Text + "]"
	}
	return sp.Text
}

func transcript(r Result) string {
	var b strings.Builder
	for _, e := range r.Effects {
		b.WriteString("! " + e.Kind.String())
		if e.Arg != "" {
			b.WriteString(" " + e.Arg)
		}
		b.WriteString("\n")
	}
	for _, bl := range r.Blocks {
		b.WriteString("<" + bl.Kind.String())
		if bl.Src != "" {
			b.WriteString(" src=" + bl.Src + " alt=" + bl.Alt)
		}
		b.WriteString(">\n")
		for _, l := range bl.Lines {
			b.WriteString("  ")
			for _, sp := range l {
				b.WriteString(renderSpan(sp))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		require.NoError(t, os.WriteFile(path, []byte(got), 0o644))
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "run go test -run TestGolden -update to create %s", path)
	assert.Equal(t, string(want), got)
}

func TestGolden(t *testing.T) {
	t.Run("the boot screen matches its transcript", func(t *testing.T) {
		golden(t, "boot.golden", transcript(newTestShell().Boot()))
	})

	t.Run("a representative session matches its transcript", func(t *testing.T) {
		s := newTestShell()
		var b strings.Builder
		for _, line := range session {
			b.WriteString("$ " + s.Prompt() + " " + line + "\n")
			b.WriteString(transcript(s.Exec(line)))
			switch line {
			case "repos":
				b.WriteString("$ <deliver repos>\n")
				b.WriteString(transcript(s.Deliver(FeedRepos, []FeedItem{
					{Name: "pgqueue", Description: "Postgres-backed durable job queue and NOTIFY/LISTEN pub/sub.", URL: "https://github.com/pascalallen/pgqueue", Meta: "Go · updated 2026-08-16"},
					{Name: "pubsub", Description: "In-memory pub/sub built on goroutines and channels.", URL: "https://github.com/pascalallen/pubsub", Meta: "Go · updated 2026-08-13"},
				})))
			case "packages":
				b.WriteString("$ <deliver error packages>\n")
				b.WriteString(transcript(s.DeliverError(FeedPackages)))
			}
		}
		golden(t, "session.golden", b.String())
	})
}
