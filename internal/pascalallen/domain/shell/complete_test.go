package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type completeCase struct {
	name   string
	setup  []string
	line   string
	cursor int
	want   Completion
}

var completeCases = []completeCase{
	{name: "a unique command completes with a space", line: "whoa", cursor: 4, want: Completion{Line: "whoami ", Cursor: 7}},
	{name: "several commands extend to the common prefix and list candidates", line: "h", cursor: 1, want: Completion{Line: "h", Cursor: 1, Candidates: []string{"head", "help", "history"}}},
	{name: "several commands with a longer common prefix extend the word", line: "pa", cursor: 2, want: Completion{Line: "packages ", Cursor: 9}},
	{name: "a shared prefix longer than the word is inserted", line: "c", cursor: 1, want: Completion{Line: "c", Cursor: 1, Candidates: []string{"camera", "cat", "cd", "clear", "contact"}}},
	{name: "hidden commands are not offered", line: "sud", cursor: 3, want: Completion{Line: "sud", Cursor: 3}},
	{name: "a command after a semicolon completes", line: "ls; pw", cursor: 6, want: Completion{Line: "ls; pwd ", Cursor: 8}},
	{name: "a command after and-and completes", line: "ls&&whoa", cursor: 8, want: Completion{Line: "ls&&whoami ", Cursor: 11}},
	{name: "a file completes with a space", line: "cat abo", cursor: 7, want: Completion{Line: "cat about.md ", Cursor: 13}},
	{name: "a directory completes with a trailing slash", line: "cat wri", cursor: 7, want: Completion{Line: "cat writing/", Cursor: 12}},
	{name: "a nested path completes", line: "cat writing/19", cursor: 14, want: Completion{Line: "cat writing/19-divergent-career-paths.md ", Cursor: 41}},
	{name: "several files extend to their common prefix", line: "cat writing/0", cursor: 13, want: Completion{Line: "cat writing/0", Cursor: 13, Candidates: []string{"01-run-a-go-web-server-with-apple-container.md", "02-event-sourcing-in-go.md", "03-live-security-camera-with-go.md", "04-structuring-go-projects-with-clean-architecture.md", "05-build-an-mcp-server-in-go.md", "06-go-ing-beyond-the-ordinary.md", "07-mastering-the-fundamentals-of-algorithms.md", "08-automate-your-deployments-with-cron.md", "09-build-a-grpc-server-in-go.md"}}},
	{name: "a common file prefix is inserted", line: "cat projects/react", cursor: 18, want: Completion{Line: "cat projects/react-", Cursor: 19, Candidates: []string{"react-event-dispatcher.md", "react-form-components.md"}}},
	{name: "an empty argument lists the directory", line: "cd ", cursor: 3, want: Completion{Line: "cd ", Cursor: 3, Candidates: []string{"experience/", "projects/", "writing/"}}},
	{name: "cd offers directories only", line: "cd e", cursor: 4, want: Completion{Line: "cd experience/", Cursor: 14}},
	{name: "cd offers no files", line: "cd abo", cursor: 6, want: Completion{Line: "cd abo", Cursor: 6}},
	{name: "a tilde path completes from home", setup: []string{"cd writing"}, line: "cat ~/con", cursor: 9, want: Completion{Line: "cat ~/contact.md ", Cursor: 17}},
	{name: "a relative path completes from the working directory", setup: []string{"cd projects"}, line: "cat pgm", cursor: 7, want: Completion{Line: "cat pgmcp.md ", Cursor: 13}},
	{name: "a parent path completes", setup: []string{"cd projects"}, line: "cat ../abo", cursor: 10, want: Completion{Line: "cat ../about.md ", Cursor: 16}},
	{name: "a bare tilde becomes home", line: "ls ~", cursor: 4, want: Completion{Line: "ls ~/", Cursor: 5}},
	{name: "man completes command names", line: "man gr", cursor: 6, want: Completion{Line: "man grep ", Cursor: 9}},
	{name: "theme completes theme names", line: "theme am", cursor: 8, want: Completion{Line: "theme amber ", Cursor: 12}},
	{name: "a mid-line cursor completes the word before it", line: "cat abo contact.md", cursor: 7, want: Completion{Line: "cat about.md contact.md", Cursor: 12}},
	{name: "a mid-line cursor inside a command keeps the rest", line: "whoa && ls", cursor: 4, want: Completion{Line: "whoami && ls", Cursor: 6}},
	{name: "no match leaves the line alone", line: "cat zzz", cursor: 7, want: Completion{Line: "cat zzz", Cursor: 7}},
	{name: "an unknown command has no argument completion", line: "nope ab", cursor: 7, want: Completion{Line: "nope ab", Cursor: 7}},
	{name: "a command without arguments offers nothing", line: "pwd a", cursor: 5, want: Completion{Line: "pwd a", Cursor: 5}},
	{name: "a flag offers nothing", line: "ls -", cursor: 4, want: Completion{Line: "ls -", Cursor: 4}},
	{name: "a missing directory offers nothing", line: "cat nope/a", cursor: 10, want: Completion{Line: "cat nope/a", Cursor: 10}},
	{name: "a cursor past the end is clamped", line: "whoa", cursor: 99, want: Completion{Line: "whoami ", Cursor: 7}},
	{name: "a negative cursor is clamped", line: "ls", cursor: -5, want: Completion{Line: "ls", Cursor: 0, Candidates: []string{"about", "camera", "cat", "cd", "clear", "contact", "date", "echo", "education", "exit", "experience", "grep", "head", "help", "history", "ls", "man", "open", "packages", "projects", "pwd", "repos", "resume", "skills", "tail", "theme", "tree", "whoami", "writing"}}},
	{name: "the cursor counts runes", line: "echo é; whoa", cursor: 12, want: Completion{Line: "echo é; whoami ", Cursor: 15}},
}

func TestComplete(t *testing.T) {
	for _, tc := range completeCases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestShell()
			for _, l := range tc.setup {
				s.Exec(l)
			}
			assert.Equal(t, tc.want, s.Complete(tc.line, tc.cursor))
		})
	}

	t.Run("the common prefix never splits a rune", func(t *testing.T) {
		assert.Equal(t, "a", commonPrefix("aé", "aè"))
		assert.Equal(t, "ab", commonPrefix("ab", "abc"))
	})
}
