package shell

import (
	"io/fs"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const historyLimit = 200

const (
	maxLineRunes    = 4096
	maxSteps        = 32
	maxOutputLines  = 2000
	maxSuggestRunes = 32
)

const defaultTheme = "dark"

var themes = []string{"dark", "light", "amber", "green"}

type Option func(*Shell)

func WithClock(now func() time.Time) Option {
	return func(s *Shell) {
		if now != nil {
			s.now = now
		}
	}
}

func WithUser(name string) Option {
	return func(s *Shell) {
		if name != "" {
			s.user = name
		}
	}
}

func WithHost(name string) Option {
	return func(s *Shell) {
		if name != "" {
			s.host = name
		}
	}
}

func WithTheme(name string) Option {
	return func(s *Shell) {
		if isTheme(name) {
			s.theme = name
		}
	}
}

type Shell struct {
	fsys     fs.FS
	now      func() time.Time
	user     string
	host     string
	theme    string
	cwd      string
	prev     string
	history  []string
	dropped  int
	commands []command
	index    map[string]int
}

func New(fsys fs.FS, opts ...Option) *Shell {
	s := &Shell{
		fsys:  fsys,
		now:   func() time.Time { return time.Unix(0, 0).UTC() },
		user:  "visitor",
		host:  "pascalallen.com",
		theme: defaultTheme,
		cwd:   ".",
		prev:  ".",
	}
	for _, opt := range opts {
		opt(s)
	}
	s.commands = registry()
	s.index = make(map[string]int, len(s.commands))
	for i, c := range s.commands {
		s.index[c.name] = i
	}
	return s
}

func (s *Shell) Prompt() string {
	return clean(s.user + "@" + s.host + ":" + tildePath(s.cwd) + "$")
}

func (s *Shell) History() []string {
	return append([]string{}, s.history...)
}

func (s *Shell) Commands() []CommandInfo {
	var infos []CommandInfo
	for _, c := range s.commands {
		if !c.hidden {
			infos = append(infos, CommandInfo{Name: c.name, Synopsis: c.synopsis, Summary: c.summary})
		}
	}
	return infos
}

func (s *Shell) Exec(line string) Result {
	o := &output{}
	if utf8.RuneCountInString(line) > maxLineRunes {
		o.fail("bash: input line too long (max " + strconv.Itoa(maxLineRunes) + " characters)")
		return o.result()
	}
	s.record(line)
	steps, err := parse(line)
	if err != nil {
		o.fail("bash: " + err.Error())
		return o.result()
	}
	if len(steps) > maxSteps {
		o.fail("bash: too many commands on one line (max " + strconv.Itoa(maxSteps) + ")")
		return o.result()
	}
	failed := false
	for _, st := range steps {
		if st.andThen && failed {
			continue
		}
		failed = s.run(o, st.words)
	}
	return o.result()
}

func (s *Shell) run(o *output, words []string) bool {
	o.failed = false
	name, args := words[0], words[1:]
	c, ok := s.lookup(name)
	if !ok {
		o.fail("bash: " + name + ": command not found")
		if guess := s.nearest(name); guess != "" {
			o.failLine(plain("did you mean "), runSpan(guess, guess), plain("?"))
		}
		return true
	}
	if c.maxArgs >= 0 && len(args) > c.maxArgs {
		o.fail(name + ": too many arguments")
		return true
	}
	c.run(s, o, args)
	return o.failed
}

func (s *Shell) lookup(name string) (command, bool) {
	i, ok := s.index[name]
	if !ok {
		return command{}, false
	}
	return s.commands[i], true
}

func (s *Shell) record(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if n := len(s.history); n > 0 && s.history[n-1] == line {
		return
	}
	s.history = append(s.history, line)
	if extra := len(s.history) - historyLimit; extra > 0 {
		s.history = append([]string{}, s.history[extra:]...)
		s.dropped += extra
	}
}

func (s *Shell) nearest(name string) string {
	n := utf8.RuneCountInString(name)
	if n > maxSuggestRunes {
		return ""
	}
	best, bestDist := "", 3
	for _, c := range s.commands {
		if c.hidden || n-len(c.name) > 2 || len(c.name)-n > 2 {
			continue
		}
		if d := distance(name, c.name); d < bestDist && d < n {
			best, bestDist = c.name, d
		}
	}
	return best
}

func distance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	d := make([][]int, len(ra)+1)
	for i := range d {
		d[i] = make([]int, len(rb)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(ra)][len(rb)]
}

func (s *Shell) Boot() Result {
	o := &output{}
	o.line(styled("pascalallen.com", StyleHeading), styled("  a read-only terminal", StyleDim))
	o.line()
	if data, err := fs.ReadFile(s.fsys, "about.md"); err == nil {
		o.lines(renderText(string(data)))
		o.line()
	}
	o.line(chips(bootCommands)...)
	o.line(styled("Type ", StyleDim), runSpan("help", "help"), styled(" for every command, or tap one above.", StyleDim))
	return o.result()
}

var bootCommands = []string{"help", "about", "resume", "experience", "projects", "writing", "contact", "repos", "camera"}

func chips(names []string) []Span {
	var spans []Span
	for i, n := range names {
		if i > 0 {
			spans = append(spans, plain("  "))
		}
		spans = append(spans, runSpan(n, n))
	}
	return spans
}

var feedTitles = map[string]string{
	FeedRepos:    "Repositories on GitHub",
	FeedPackages: "Packages on npm",
}

func (s *Shell) Deliver(feed string, items []FeedItem) Result {
	o := &output{}
	title, ok := feedTitles[feed]
	if !ok {
		o.fail(feed + ": unknown feed")
		return o.result()
	}
	o.line(styled(title, StyleHeading))
	if len(items) == 0 {
		o.line(styled("nothing to show yet", StyleDim))
	}
	for _, it := range items {
		name := plain(it.Name)
		if isWebURL(it.URL) {
			name = linkSpan(it.Name, it.URL)
		}
		spans := []Span{name}
		if it.Meta != "" {
			spans = append(spans, styled("  "+it.Meta, StyleDim))
		}
		o.line(spans...)
		if it.Description != "" {
			o.line(plain("  " + it.Description))
		}
	}
	return o.result()
}

func (s *Shell) DeliverError(feed string) Result {
	o := &output{}
	if _, ok := feedTitles[feed]; !ok {
		o.fail(feed + ": unknown feed")
		return o.result()
	}
	o.fail(feed + ": could not load the list right now")
	o.failLine(plain("try again: "), runSpan(feed, feed))
	return o.result()
}

func isTheme(name string) bool {
	for _, t := range themes {
		if t == name {
			return true
		}
	}
	return false
}
