package shell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type commandCase struct {
	name    string
	setup   []string
	line    string
	want    []string
	notWant []string
	failed  bool
	effects []Effect
	runs    []string
	hrefs   []string
}

const okMarker = "__ok__"

var commandCases = map[string][]commandCase{
	"help": {
		{name: "lists every visible command with its summary", line: "help", want: []string{"ls", "list directory contents", "camera"}, notWant: []string{"sudo", "mkdir"}, runs: []string{"ls", "man ls"}},
		{name: "shows the manual page for one command", line: "help cat", want: []string{"NAME", "cat - print files", "SYNOPSIS"}},
		{name: "rejects an unknown topic", line: "help nope", want: []string{"help: no help topics match 'nope'"}, failed: true},
	},
	"man": {
		{name: "prints name, synopsis and description", line: "man ls", want: []string{"NAME", "    ls - list directory contents", "SYNOPSIS", "ls [-la] [path]", "DESCRIPTION", "-l  long format"}},
		{name: "asks which page is wanted", line: "man", want: []string{"What manual page do you want?"}, failed: true, runs: []string{"man ls"}},
		{name: "reports a missing page", line: "man nope", want: []string{"No manual entry for nope"}, failed: true},
	},
	"ls": {
		{name: "lists the home directory", line: "ls", want: []string{"about.md", "experience/", "writing/", "README.md"}, runs: []string{"cat ~/about.md", "ls ~/writing"}},
		{name: "lists a subdirectory", line: "ls projects", want: []string{"pgqueue.md"}, runs: []string{"cat ~/projects/pgqueue.md"}},
		{name: "lists in long format with dot entries", line: "ls -la", want: []string{"dr-xr-xr-x  visitor       -  ./", "../", "-r--r--r--  visitor", "about.md"}},
		{name: "shows dot entries in short format", line: "ls -a writing", want: []string{"./  ../  01-"}},
		{name: "lists a single file", line: "ls about.md", want: []string{"about.md"}, notWant: []string{"contact.md"}},
		{name: "lists a single file in long format", line: "ls -l about.md", want: []string{"-r--r--r--  visitor", "about.md"}},
		{name: "reports a missing path", line: "ls nope", want: []string{"ls: cannot access 'nope': No such file or directory"}, failed: true},
		{name: "rejects an unknown option", line: "ls -z", want: []string{"ls: invalid option -- 'z'"}, failed: true},
		{name: "rejects two paths", line: "ls writing projects", want: []string{"ls: too many arguments"}, failed: true},
	},
	"cd": {
		{name: "changes into a directory", line: "cd writing && pwd", want: []string{"/home/visitor/writing"}},
		{name: "goes home with no argument", setup: []string{"cd writing"}, line: "cd && pwd", want: []string{"/home/visitor"}, notWant: []string{"writing"}},
		{name: "returns to the previous directory", setup: []string{"cd writing", "cd ~/projects"}, line: "cd - && pwd", want: []string{"/home/visitor/writing"}},
		{name: "accepts the absolute home path", line: "cd /home/visitor/projects && pwd", want: []string{"/home/visitor/projects"}},
		{name: "reports a missing directory", line: "cd nope", want: []string{"cd: nope: No such file or directory"}, failed: true},
		{name: "refuses a file", line: "cd about.md", want: []string{"cd: about.md: Not a directory"}, failed: true},
		{name: "rejects two arguments", line: "cd a b", want: []string{"cd: too many arguments"}, failed: true},
	},
	"pwd": {
		{name: "prints the home directory", line: "pwd", want: []string{"/home/visitor"}},
		{name: "rejects arguments", line: "pwd x", want: []string{"pwd: too many arguments"}, failed: true},
	},
	"cat": {
		{name: "prints a file", line: "cat about.md", want: []string{"Pascal Allen", "Full-stack software engineer"}},
		{name: "prints several files in order", line: "cat about.md contact.md", want: []string{"Full-stack", "pascal.allen88@gmail.com"}, hrefs: []string{"mailto:pascal.allen88@gmail.com", "https://github.com/pascalallen"}},
		{name: "keeps printing after a missing file and fails", line: "cat nope about.md", want: []string{"cat: nope: No such file or directory", "Pascal Allen"}, failed: true},
		{name: "asks for a file", line: "cat", want: []string{"cat: missing file operand"}, failed: true},
		{name: "refuses a directory", line: "cat writing", want: []string{"cat: writing: Is a directory"}, failed: true},
	},
	"tree": {
		{name: "draws the whole tree with totals", line: "tree", want: []string{"~", "├── about.md", "├── experience/", "│   ├── 01-sectigo.md", "└── writing/", "3 directories, 44 files"}},
		{name: "draws a subtree", line: "tree experience", want: []string{"~/experience", "└── 06-crimson-drive-design.md", "0 directories, 6 files"}},
		{name: "reports a missing path", line: "tree nope", want: []string{"tree: nope: No such file or directory"}, failed: true},
		{name: "refuses a file", line: "tree about.md", want: []string{"tree: about.md: Not a directory"}, failed: true},
	},
	"head": {
		{name: "prints the first line", line: "head -n 1 about.md", want: []string{"Pascal Allen"}, notWant: []string{"Full-stack"}},
		{name: "accepts the short count form", line: "head -3 about.md", want: []string{"Senior Software Engineer"}, notWant: []string{"Full-stack"}},
		{name: "accepts the attached count form", line: "head -n2 about.md", want: []string{"Pascal Allen"}, notWant: []string{"Senior"}},
		{name: "labels several files", line: "head -n 1 about.md contact.md", want: []string{"==> about.md <==", "==> contact.md <==", "Contact"}},
		{name: "prints ten lines by default", line: "head resume.md", want: []string{"Pascal Allen"}, notWant: []string{"Summary"}},
		{name: "asks for a file", line: "head", want: []string{"head: missing file operand"}, failed: true},
		{name: "rejects a bad count", line: "head -n x about.md", want: []string{"head: invalid number of lines: 'x'"}, failed: true},
		{name: "rejects a missing count", line: "head about.md -n", want: []string{"head: option requires an argument -- 'n'"}, failed: true},
		{name: "rejects an unknown option", line: "head -q about.md", want: []string{"head: invalid option -- 'q'"}, failed: true},
		{name: "reports a missing file", line: "head nope", want: []string{"head: nope: No such file or directory"}, failed: true},
	},
	"tail": {
		{name: "prints the last line", line: "tail -n 1 contact.md", want: []string{"LinkedIn"}, notWant: []string{"Email"}},
		{name: "prints nothing for zero lines", line: "tail -n 0 contact.md", notWant: []string{"LinkedIn"}},
		{name: "reports a missing file", line: "tail nope", want: []string{"tail: nope: No such file or directory"}, failed: true},
		{name: "rejects a negative count", line: "tail -n -1 about.md", want: []string{"tail: invalid number of lines: '-1'"}, failed: true},
	},
	"grep": {
		{name: "finds text recursively with file and line", line: "grep Sectigo", want: []string{"experience/01-sectigo.md:1: # Sectigo"}, runs: []string{"cat ~/experience/01-sectigo.md"}},
		{name: "ignores case with -i", line: "grep -i SECTIGO experience", want: []string{"experience/01-sectigo.md:1:"}},
		{name: "matches case by default", line: "grep SECTIGO experience", failed: true},
		{name: "searches a single file", line: "grep Medium contact.md", want: []string{"contact.md:6: Medium"}},
		{name: "fails quietly without matches", line: "grep zzqqxx", failed: true},
		{name: "asks for a pattern", line: "grep", want: []string{"usage: grep [-i] <text> [path]"}, failed: true},
		{name: "reports a missing path", line: "grep Go nope", want: []string{"grep: nope: No such file or directory"}, failed: true},
		{name: "rejects an unknown option", line: "grep -z Go", want: []string{"grep: invalid option -- 'z'"}, failed: true},
		{name: "rejects extra paths", line: "grep Go a b", want: []string{"grep: too many arguments"}, failed: true},
	},
	"whoami": {
		{name: "prints the user", line: "whoami", want: []string{"visitor"}},
		{name: "rejects arguments", line: "whoami me", want: []string{"whoami: too many arguments"}, failed: true},
	},
	"echo": {
		{name: "prints its arguments", line: `echo hello   "big world"`, want: []string{"hello big world"}},
		{name: "reports an unterminated quote", line: `echo "abc`, want: []string{"bash: unexpected EOF while looking for matching `\"'"}, failed: true},
	},
	"date": {
		{name: "prints the injected clock", line: "date", want: []string{"Wed Nov 11 09:30:00 UTC 2026"}},
		{name: "rejects arguments", line: "date +%s", want: []string{"date: too many arguments"}, failed: true},
	},
	"history": {
		{name: "numbers previous commands", setup: []string{"ls", "pwd"}, line: "history", want: []string{"    1  ls", "    2  pwd"}, runs: []string{"ls", "pwd"}},
		{name: "limits to the last n", setup: []string{"ls", "pwd"}, line: "history 1", want: []string{"history 1"}, notWant: []string{"  ls"}},
		{name: "rejects a non-number", line: "history x", want: []string{"history: x: numeric argument required"}, failed: true},
	},
	"clear": {
		{name: "emits a clear effect", line: "clear", effects: []Effect{{Kind: EffectClear}}},
		{name: "rejects arguments", line: "clear now", want: []string{"clear: too many arguments"}, failed: true},
	},
	"open": {
		{name: "opens a url", line: "open https://github.com/pascalallen", want: []string{"opening https://github.com/pascalallen"}, effects: []Effect{{Kind: EffectOpen, Arg: "https://github.com/pascalallen"}}},
		{name: "opens a mailto link", line: "open mailto:pascal.allen88@gmail.com", effects: []Effect{{Kind: EffectOpen, Arg: "mailto:pascal.allen88@gmail.com"}}},
		{name: "opens the first web link in a file", line: "open writing/19-divergent-career-paths.md", effects: []Effect{{Kind: EffectOpen, Arg: "https://www.bizjournals.com/sanantonio/news/2016/11/23/divergent-career-paths-how-tech-talent-is-leaking.html"}}},
		{name: "skips the email to open the first web link", line: "open contact.md", effects: []Effect{{Kind: EffectOpen, Arg: "https://pascalallen.com"}}},
		{name: "asks for an operand", line: "open", want: []string{"open: missing operand"}, failed: true},
		{name: "refuses a javascript url", line: "open javascript:alert(1)", want: []string{"open: javascript:alert(1): No such file or directory"}, failed: true, effects: []Effect{}},
		{name: "reports a file without links", line: "open education.md", want: []string{"open: education.md: no link found"}, failed: true},
		{name: "refuses a directory", line: "open writing", want: []string{"open: writing: Is a directory"}, failed: true},
	},
	"about": {
		{name: "prints the bio with more commands", line: "about", want: []string{"Pascal Allen", "Full-stack", "more: resume"}, runs: []string{"resume", "contact"}},
		{name: "rejects arguments", line: "about me", want: []string{"about: too many arguments"}, failed: true},
	},
	"resume": {
		{name: "prints the full resume", line: "resume", want: []string{"Summary", "Professional Experience", "Education & Certifications"}},
		{name: "rejects arguments", line: "resume pdf", want: []string{"resume: too many arguments"}, failed: true},
	},
	"skills": {
		{name: "prints the skills", line: "skills", want: []string{"Languages", "Go, PHP, React"}},
		{name: "rejects arguments", line: "skills go", want: []string{"skills: too many arguments"}, failed: true},
	},
	"experience": {
		{name: "lists employers with dates and drill-in commands", line: "experience", want: []string{"Experience", "Sectigo — Senior Software Engineer", "Austin, TX · 2022 – 2024"}, runs: []string{"cat ~/experience/01-sectigo.md", "cat ~/experience/06-crimson-drive-design.md"}},
		{name: "rejects arguments", line: "experience all", want: []string{"experience: too many arguments"}, failed: true},
	},
	"projects": {
		{name: "lists projects with drill-in commands", line: "projects", want: []string{"Projects", "pgqueue (v1.1)"}, runs: []string{"cat ~/projects/pgqueue.md", "ls ~/projects"}},
		{name: "rejects arguments", line: "projects go", want: []string{"projects: too many arguments"}, failed: true},
	},
	"writing": {
		{name: "lists publications with drill-in commands", line: "writing", want: []string{"Writing", "How To: Run a Go Web Server With Apple Container", "Go-ing Beyond the Ordinary: Discovering the Unique Superpowers of the Go Language"}, runs: []string{"cat ~/writing/19-divergent-career-paths.md"}},
		{name: "rejects arguments", line: "writing go", want: []string{"writing: too many arguments"}, failed: true},
	},
	"contact": {
		{name: "prints the links", line: "contact", want: []string{"Email"}, hrefs: []string{"mailto:pascal.allen88@gmail.com", "https://www.linkedin.com/in/pascal-allen-942749112"}},
		{name: "rejects arguments", line: "contact me", want: []string{"contact: too many arguments"}, failed: true},
	},
	"education": {
		{name: "prints the certifications", line: "education", want: []string{"Harvard Online — CS50"}},
		{name: "rejects arguments", line: "education x", want: []string{"education: too many arguments"}, failed: true},
	},
	"repos": {
		{name: "emits a fetch effect and a loading line", line: "repos", want: []string{"fetching repositories from GitHub..."}, effects: []Effect{{Kind: EffectFetch, Arg: "repos"}}},
		{name: "rejects arguments", line: "repos all", want: []string{"repos: too many arguments"}, failed: true},
	},
	"packages": {
		{name: "emits a fetch effect and a loading line", line: "packages", want: []string{"fetching packages from npm..."}, effects: []Effect{{Kind: EffectFetch, Arg: "packages"}}},
		{name: "rejects arguments", line: "packages all", want: []string{"packages: too many arguments"}, failed: true},
	},
	"camera": {
		{name: "emits a media block for the stream", line: "camera", want: []string{"live from a Raspberry Pi"}},
		{name: "rejects arguments", line: "camera on", want: []string{"camera: too many arguments"}, failed: true},
	},
	"theme": {
		{name: "lists themes and marks the current one", line: "theme", want: []string{"* dark", "  light", "  amber", "  green"}, runs: []string{"theme light"}},
		{name: "sets a theme", line: "theme amber", want: []string{"theme set to amber"}, effects: []Effect{{Kind: EffectTheme, Arg: "amber"}}},
		{name: "remembers the new theme", setup: []string{"theme green"}, line: "theme", want: []string{"* green", "  dark"}},
		{name: "rejects an unknown theme", line: "theme pink", want: []string{"theme: pink: unknown theme (try dark, light, amber, green)"}, failed: true},
	},
	"exit": {
		{name: "logs out politely", line: "exit", want: []string{"logout"}},
		{name: "rejects arguments", line: "exit 1", want: []string{"exit: too many arguments"}, failed: true},
	},
	"sudo": {
		{name: "refuses root", line: "sudo ls", want: []string{"visitor is not in the sudoers file. This incident will be reported."}, failed: true},
	},
	"vim": {
		{name: "explains there is no editor", line: "vim about.md", want: []string{"vim: there is no editor"}},
		{name: "rejects two files", line: "vim a b", want: []string{"vim: too many arguments"}, failed: true},
	},
	"rm": {
		{name: "refuses to remove a file", line: "rm about.md", want: []string{"rm: cannot remove 'about.md': Read-only file system"}, failed: true},
		{name: "has a joke for rm -rf /", line: "rm -rf /", want: []string{"rm: nice try"}, failed: true},
		{name: "asks for an operand", line: "rm -f", want: []string{"rm: missing operand"}, failed: true},
		{name: "rejects an unknown option", line: "rm -z x", want: []string{"rm: invalid option -- 'z'"}, failed: true},
	},
	"mkdir": {
		{name: "refuses to create a directory", line: "mkdir notes", want: []string{"mkdir: cannot create directory 'notes': Read-only file system"}, failed: true},
		{name: "asks for an operand", line: "mkdir", want: []string{"mkdir: missing operand"}, failed: true},
	},
	"touch": {
		{name: "refuses to create a file", line: "touch x", want: []string{"touch: cannot touch 'x': Read-only file system"}, failed: true},
	},
	"mv": {
		{name: "refuses to move a file", line: "mv about.md bio.md", want: []string{"mv: cannot move 'about.md' to 'bio.md': Read-only file system"}, failed: true},
		{name: "asks for two operands", line: "mv about.md", want: []string{"mv: missing file operand"}, failed: true},
	},
}

var alwaysFails = map[string]bool{"sudo": true, "rm": true, "mkdir": true, "touch": true, "mv": true}

func runCase(t *testing.T, tc commandCase) (Result, bool) {
	t.Helper()
	s := newTestShell()
	for _, l := range tc.setup {
		s.Exec(l)
	}
	r := s.Exec(tc.line + " && echo " + okMarker)
	if n := len(r.Blocks); n > 0 {
		last := &r.Blocks[n-1]
		if k := len(last.Lines); k > 0 && lineText(last.Lines[k-1]) == okMarker {
			last.Lines = last.Lines[:k-1]
			if len(last.Lines) == 0 {
				r.Blocks = r.Blocks[:n-1]
			}
			return r, true
		}
	}
	return r, false
}

func TestCommands(t *testing.T) {
	for name, cases := range commandCases {
		for _, tc := range cases {
			t.Run(name+" "+tc.name, func(t *testing.T) {
				r, ok := runCase(t, tc)
				text := resultText(r)
				assert.Equal(t, !tc.failed, ok, text)
				for _, w := range tc.want {
					assert.Contains(t, text, w)
				}
				for _, w := range tc.notWant {
					assert.NotContains(t, text, w)
				}
				if tc.effects != nil {
					want := tc.effects
					if len(want) == 0 {
						want = nil
					}
					assert.Equal(t, want, r.Effects)
				}
				for _, run := range tc.runs {
					assert.Contains(t, runs(r), run)
				}
				for _, href := range tc.hrefs {
					assert.Contains(t, hrefs(r), href)
				}
				if tc.failed && len(tc.want) > 0 {
					assert.NotEmpty(t, errorText(r))
				}
			})
		}
	}
}

func TestCommandRegistry(t *testing.T) {
	s := newTestShell()

	t.Run("every registered command has help text, a man page and table cases", func(t *testing.T) {
		for _, c := range s.commands {
			assert.NotEmpty(t, c.synopsis, c.name)
			assert.NotEmpty(t, c.summary, c.name)
			assert.NotEmpty(t, c.help, c.name)
			assert.True(t, strings.HasPrefix(c.synopsis, c.name), c.name)
			man := s.Exec("man " + c.name)
			assert.Contains(t, resultText(man), c.name+" - "+c.summary)
			assert.Empty(t, errorText(man), c.name)
			cases := commandCases[c.name]
			require.NotEmpty(t, cases, c.name+" has no table cases")
			var passes, fails int
			for _, tc := range cases {
				if tc.failed {
					fails++
				} else {
					passes++
				}
			}
			assert.Positive(t, fails, c.name+" has no error-path case")
			if !alwaysFails[c.name] {
				assert.Positive(t, passes, c.name+" has no happy-path case")
			}
		}
	})

	t.Run("every table entry names a registered command", func(t *testing.T) {
		for name := range commandCases {
			_, ok := s.lookup(name)
			assert.True(t, ok, name)
		}
	})

	t.Run("command names are unique", func(t *testing.T) {
		assert.Len(t, s.index, len(s.commands))
	})

	t.Run("Commands lists the visible commands with synopsis and summary", func(t *testing.T) {
		infos := s.Commands()
		names := map[string]bool{}
		for _, info := range infos {
			names[info.Name] = true
			assert.NotEmpty(t, info.Synopsis)
			assert.NotEmpty(t, info.Summary)
		}
		for _, want := range []string{"help", "man", "ls", "cd", "pwd", "cat", "tree", "head", "tail", "grep", "whoami", "echo", "date", "history", "clear", "open", "about", "resume", "skills", "experience", "projects", "writing", "contact", "education", "repos", "packages", "camera", "theme", "exit"} {
			assert.True(t, names[want], want)
		}
		for _, hidden := range []string{"sudo", "vim", "rm", "mkdir", "touch", "mv"} {
			assert.False(t, names[hidden], hidden)
		}
	})
}
