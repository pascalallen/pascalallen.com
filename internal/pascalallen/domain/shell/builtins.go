package shell

import (
	"io/fs"
	"strconv"
	"strings"
	"time"
)

const cameraStream = "/api/v1/camera/stream"

func splitFlags(args []string, allowed string) (map[rune]bool, []string, string) {
	flags := map[rune]bool{}
	var rest []string
	for i, a := range args {
		if a == "--" {
			rest = append(rest, args[i+1:]...)
			break
		}
		if len(a) < 2 || a[0] != '-' {
			rest = append(rest, a)
			continue
		}
		for _, r := range a[1:] {
			if !strings.ContainsRune(allowed, r) {
				return nil, nil, string(r)
			}
			flags[r] = true
		}
	}
	return flags, rest, ""
}

func cmdHelp(s *Shell, o *output, args []string) {
	if len(args) == 1 {
		c, ok := s.lookup(args[0])
		if !ok {
			o.fail("help: no help topics match '" + args[0] + "'")
			return
		}
		manPage(o, c)
		return
	}
	o.line(styled("Commands", StyleHeading))
	for _, c := range s.commands {
		if c.hidden {
			continue
		}
		o.line(plain("  "), runSpan(c.name, c.name), plain(strings.Repeat(" ", max(1, 12-len(c.name)))+c.summary))
	}
	o.line()
	o.line(styled("Type ", StyleDim), runSpan("man ls", "man ls"), styled(" for details. Tab completes; tap any command to run it.", StyleDim))
}

func cmdMan(s *Shell, o *output, args []string) {
	if len(args) == 0 {
		o.fail("What manual page do you want?")
		o.failLine(plain("For example, try "), runSpan("man ls", "man ls"))
		return
	}
	c, ok := s.lookup(args[0])
	if !ok {
		o.fail("No manual entry for " + args[0])
		return
	}
	manPage(o, c)
}

func manPage(o *output, c command) {
	o.line(styled("NAME", StyleHeading))
	o.line(plain("    " + c.name + " - " + c.summary))
	o.line()
	o.line(styled("SYNOPSIS", StyleHeading))
	o.line(plain("    "), styled(c.synopsis, StyleBold))
	o.line()
	o.line(styled("DESCRIPTION", StyleHeading))
	for _, l := range strings.Split(c.help, "\n") {
		o.line(plain("    " + l))
	}
}

func cmdLs(s *Shell, o *output, args []string) {
	flags, rest, bad := splitFlags(args, "la")
	if bad != "" {
		o.fail("ls: invalid option -- '" + bad + "'")
		return
	}
	if len(rest) > 1 {
		o.fail("ls: too many arguments")
		return
	}
	arg := ""
	if len(rest) == 1 {
		arg = rest[0]
	}
	p, info, problem := s.locate(arg)
	if problem != "" {
		o.fail("ls: cannot access '" + arg + "': " + problem)
		return
	}
	if !info.IsDir() {
		if flags['l'] {
			o.line(longEntry(s, p, info)...)
		} else {
			o.line(entrySpan(p, info.Name(), false))
		}
		return
	}
	entries, err := fs.ReadDir(s.fsys, p)
	if err != nil {
		o.fail("ls: cannot open directory '" + arg + "'")
		return
	}
	if flags['l'] {
		if flags['a'] {
			o.line(plain("dr-xr-xr-x  "+s.user+"       -  "), styled("./", StylePath))
			o.line(plain("dr-xr-xr-x  "+s.user+"       -  "), styled("../", StylePath))
		}
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				continue
			}
			o.line(longEntry(s, join(p, e.Name()), info)...)
		}
		return
	}
	var spans []Span
	if flags['a'] {
		spans = append(spans, styled("./", StylePath), plain("  "), styled("../", StylePath))
	}
	for _, e := range entries {
		if len(spans) > 0 {
			spans = append(spans, plain("  "))
		}
		spans = append(spans, entrySpan(join(p, e.Name()), e.Name(), e.IsDir()))
	}
	o.line(spans...)
}

func longEntry(s *Shell, p string, info fs.FileInfo) []Span {
	mode, size := "-r--r--r--", strconv.FormatInt(info.Size(), 10)
	if info.IsDir() {
		mode, size = "dr-xr-xr-x", "-"
	}
	pad := strings.Repeat(" ", max(0, 6-len(size)))
	return []Span{plain(mode + "  " + s.user + "  " + pad + size + "  "), entrySpan(p, info.Name(), info.IsDir())}
}

func cmdCd(s *Shell, o *output, args []string) {
	arg, label := "~", "~"
	if len(args) == 1 {
		arg, label = args[0], args[0]
		if arg == "-" {
			arg = "~/" + s.prev
		}
	}
	target, info, problem := s.locate(arg)
	if problem == "" && !info.IsDir() {
		problem = problemNotDir
	}
	if problem != "" {
		o.fail("cd: " + label + ": " + problem)
		return
	}
	s.prev, s.cwd = s.cwd, target
}

func cmdPwd(s *Shell, o *output, _ []string) {
	o.line(plain(s.displayPath(s.cwd)))
}

func cmdCat(s *Shell, o *output, args []string) {
	if len(args) == 0 {
		o.fail("cat: missing file operand")
		return
	}
	for _, a := range args {
		text, problem := s.read(a)
		if problem != "" {
			o.fail("cat: " + a + ": " + problem)
			continue
		}
		o.lines(renderText(text))
	}
}

func cmdTree(s *Shell, o *output, args []string) {
	arg := ""
	if len(args) == 1 {
		arg = args[0]
	}
	p, info, problem := s.locate(arg)
	if problem == "" && !info.IsDir() {
		problem = problemNotDir
	}
	if problem != "" {
		o.fail("tree: " + arg + ": " + problem)
		return
	}
	o.line(Span{Text: tildePath(p), Style: StylePath, Run: "ls " + shellQuote(tildePath(p))})
	dirs, files := s.walkTree(o, p, "")
	o.line()
	o.line(plain(strconv.Itoa(dirs) + " directories, " + strconv.Itoa(files) + " files"))
}

func (s *Shell) walkTree(o *output, p, prefix string) (int, int) {
	entries, _ := fs.ReadDir(s.fsys, p)
	dirs, files := 0, 0
	for i, e := range entries {
		branch, indent := "├── ", "│   "
		if i == len(entries)-1 {
			branch, indent = "└── ", "    "
		}
		child := join(p, e.Name())
		o.line(styled(prefix+branch, StyleDim), entrySpan(child, e.Name(), e.IsDir()))
		if e.IsDir() {
			d, f := s.walkTree(o, child, prefix+indent)
			dirs, files = dirs+d+1, files+f
		} else {
			files++
		}
	}
	return dirs, files
}

func headTail(name string, fromEnd bool) func(s *Shell, o *output, args []string) {
	return func(s *Shell, o *output, args []string) {
		n, sign := 10, byte(0)
		var files []string
		for i := 0; i < len(args); i++ {
			a := args[i]
			value := ""
			switch {
			case a == "-n":
				if i+1 >= len(args) {
					o.fail(name + ": option requires an argument -- 'n'")
					return
				}
				i++
				value = args[i]
			case strings.HasPrefix(a, "-n"):
				value = a[2:]
			case len(a) > 1 && a[0] == '-' && a[1] >= '0' && a[1] <= '9':
				value = a[1:]
			case len(a) > 1 && a[0] == '-':
				o.fail(name + ": invalid option -- '" + string([]rune(a[1:])[0]) + "'")
				return
			default:
				files = append(files, a)
				continue
			}
			v, vsign, ok := parseCount(value)
			if !ok {
				o.fail(name + ": invalid number of lines: '" + value + "'")
				return
			}
			n, sign = v, vsign
		}
		if len(files) == 0 {
			o.fail(name + ": missing file operand")
			return
		}
		for i, f := range files {
			text, problem := s.read(f)
			if problem != "" {
				o.fail(name + ": " + f + ": " + problem)
				continue
			}
			if len(files) > 1 {
				if i > 0 {
					o.line()
				}
				o.line(styled("==> "+f+" <==", StyleBold))
			}
			lines := renderText(text)
			switch {
			case fromEnd && sign == '+':
				o.lines(lines[min(len(lines), max(n, 1)-1):])
			case fromEnd:
				o.lines(lines[max(0, len(lines)-n):])
			case sign == '-':
				o.lines(lines[:max(0, len(lines)-n)])
			default:
				o.lines(lines[:min(n, len(lines))])
			}
		}
	}
}

func parseCount(value string) (int, byte, bool) {
	sign := byte(0)
	if value != "" && (value[0] == '+' || value[0] == '-') {
		sign, value = value[0], value[1:]
	}
	if strings.ContainsAny(value, "+-") {
		return 0, 0, false
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, 0, false
	}
	return v, sign, true
}

func cmdGrep(s *Shell, o *output, args []string) {
	flags, rest, bad := splitFlags(args, "irR")
	if bad != "" {
		o.fail("grep: invalid option -- '" + bad + "'")
		return
	}
	if len(rest) == 0 {
		o.fail("usage: grep [-i] <text> [path]")
		return
	}
	if len(rest) > 2 {
		o.fail("grep: too many arguments")
		return
	}
	pattern, arg := rest[0], ""
	if len(rest) == 2 {
		arg = rest[1]
	}
	root, _, problem := s.locate(arg)
	if problem != "" {
		o.fail("grep: " + arg + ": " + problem)
		return
	}
	fold := flags['i']
	needle := pattern
	if fold {
		needle = strings.ToLower(pattern)
	}
	found := false
	_ = fs.WalkDir(s.fsys, root, func(p string, _ fs.DirEntry, _ error) error {
		data, err := fs.ReadFile(s.fsys, p)
		if err != nil {
			return nil
		}
		for n, l := range strings.Split(strings.TrimSuffix(string(data), "\n"), "\n") {
			hay := l
			if fold {
				hay = strings.ToLower(l)
			}
			if !strings.Contains(hay, needle) {
				continue
			}
			found = true
			spans := []Span{
				{Text: p, Style: StylePath, Run: "cat " + shellQuote(tildePath(p))},
				styled(":"+strconv.Itoa(n+1)+":", StyleDim),
				plain(" "),
			}
			o.line(append(spans, highlight(l, hay, needle)...)...)
		}
		return nil
	})
	if !found {
		o.failed = true
	}
}

func highlight(line, hay, needle string) []Span {
	if needle == "" || len(hay) != len(line) {
		return []Span{plain(line)}
	}
	var spans []Span
	for {
		i := strings.Index(hay, needle)
		if i < 0 {
			break
		}
		if i > 0 {
			spans = append(spans, plain(line[:i]))
		}
		spans = append(spans, styled(line[i:i+len(needle)], StyleWarning))
		line, hay = line[i+len(needle):], hay[i+len(needle):]
	}
	if line != "" {
		spans = append(spans, plain(line))
	}
	return spans
}

func cmdWhoami(s *Shell, o *output, _ []string) {
	o.line(plain(s.user))
}

func cmdEcho(_ *Shell, o *output, args []string) {
	o.line(plain(strings.Join(args, " ")))
}

func cmdDate(s *Shell, o *output, _ []string) {
	o.line(plain(s.now().Format(time.UnixDate)))
}

func cmdHistory(s *Shell, o *output, args []string) {
	n := len(s.history)
	if len(args) == 1 {
		v, err := strconv.Atoi(args[0])
		if err != nil || v < 0 {
			o.fail("history: " + args[0] + ": numeric argument required")
			return
		}
		n = min(v, n)
	}
	for i := len(s.history) - n; i < len(s.history); i++ {
		num := strconv.Itoa(s.dropped + i + 1)
		o.line(styled(strings.Repeat(" ", max(0, 5-len(num)))+num+"  ", StyleDim), runSpan(s.history[i], s.history[i]))
	}
}

func cmdClear(_ *Shell, o *output, _ []string) {
	o.clear()
}

func cmdOpen(s *Shell, o *output, args []string) {
	if len(args) == 0 {
		o.fail("open: missing operand")
		return
	}
	arg := args[0]
	if isWebURL(arg) || isMailto(arg) {
		o.line(plain("opening "), linkSpan(arg, arg))
		o.effect(EffectOpen, arg)
		return
	}
	text, problem := s.read(arg)
	if problem != "" {
		o.fail("open: " + arg + ": " + problem)
		return
	}
	url := firstURL(renderText(text))
	if url == "" {
		o.fail("open: " + arg + ": no link found")
		return
	}
	o.line(plain("opening "), linkSpan(url, url))
	o.effect(EffectOpen, url)
}

var navCommands = []string{"about", "resume", "experience", "projects", "writing", "contact"}

func navRow(o *output, exclude string) {
	var names []string
	for _, n := range navCommands {
		if n != exclude {
			names = append(names, n)
		}
	}
	o.line()
	o.line(append([]Span{styled("more: ", StyleDim)}, chips(names)...)...)
}

func showFile(name, file string) func(s *Shell, o *output, args []string) {
	return func(s *Shell, o *output, _ []string) {
		data, err := fs.ReadFile(s.fsys, file)
		if err != nil {
			o.fail(name + ": ~/" + file + ": " + problemMissing)
			return
		}
		o.lines(renderText(string(data)))
		navRow(o, name)
	}
}

func listing(dir, title string, detail bool) func(s *Shell, o *output, args []string) {
	return func(s *Shell, o *output, _ []string) {
		entries, err := fs.ReadDir(s.fsys, dir)
		if err != nil {
			o.fail(dir + ": ~/" + dir + ": " + problemMissing)
			return
		}
		o.line(styled(title, StyleHeading))
		for _, e := range entries {
			p := dir + "/" + e.Name()
			data, err := fs.ReadFile(s.fsys, p)
			if err != nil {
				continue
			}
			name, sub := summarize(string(data))
			if name == "" {
				name = e.Name()
			}
			o.line(plain("  "), runSpan(name, "cat "+shellQuote(tildePath(p))))
			if detail && sub != "" {
				o.line(styled("    "+sub, StyleDim))
			}
		}
		o.line()
		o.line(styled("Tap an entry, or run ", StyleDim), runSpan("ls ~/"+dir, "ls ~/"+dir), styled(" and cat a file.", StyleDim))
		navRow(o, dir)
	}
}

func fetchFeed(feed, message string) func(s *Shell, o *output, args []string) {
	return func(_ *Shell, o *output, _ []string) {
		o.line(styled(message, StyleDim))
		o.effect(EffectFetch, feed)
	}
}

func cmdCamera(_ *Shell, o *output, _ []string) {
	o.media(cameraStream, "Live camera stream from a Raspberry Pi", Line{styled("live from a Raspberry Pi; type clear to close it", StyleDim)})
}

func cmdTheme(s *Shell, o *output, args []string) {
	if len(args) == 0 {
		for _, t := range themes {
			mark := "  "
			if t == s.theme {
				mark = "* "
			}
			o.line(plain(mark), runSpan(t, "theme "+t))
		}
		return
	}
	if !isTheme(args[0]) {
		o.fail("theme: " + args[0] + ": unknown theme (try " + strings.Join(themes, ", ") + ")")
		return
	}
	s.theme = args[0]
	o.line(plain("theme set to "), styled(args[0], StyleSuccess))
	o.effect(EffectTheme, args[0])
}

func cmdExit(_ *Shell, o *output, _ []string) {
	o.line(plain("logout"))
	o.line(styled("This shell lives in a browser tab, so it stays open. Close the tab to leave, or type ", StyleDim), runSpan("help", "help"), styled(".", StyleDim))
}

func cmdSudo(s *Shell, o *output, _ []string) {
	o.fail(s.user + " is not in the sudoers file. This incident will be reported.")
}

func cmdVim(_ *Shell, o *output, _ []string) {
	o.line(plain("vim: there is no editor on a read-only file system. Use "), runSpan("cat", "help cat"), plain(" to read files."))
	o.line(styled("(and you will never need to remember how to quit)", StyleDim))
}

func cmdRm(_ *Shell, o *output, args []string) {
	flags, rest, bad := splitFlags(args, "rRfiv")
	if bad != "" {
		o.fail("rm: invalid option -- '" + bad + "'")
		return
	}
	if len(rest) == 0 {
		o.fail("rm: missing operand")
		return
	}
	for _, a := range rest {
		if (flags['r'] || flags['R']) && (a == "/" || a == "~" || a == "/home") {
			o.fail("rm: nice try. Everything here is read-only, including the root.")
			return
		}
	}
	for _, a := range rest {
		o.fail("rm: cannot remove '" + a + "': " + problemReadOnly)
	}
}

func readOnly(name, verb string) func(s *Shell, o *output, args []string) {
	return func(_ *Shell, o *output, args []string) {
		if len(args) == 0 {
			o.fail(name + ": missing operand")
			return
		}
		for _, a := range args {
			o.fail(name + ": " + verb + " '" + a + "': " + problemReadOnly)
		}
	}
}

func cmdMv(_ *Shell, o *output, args []string) {
	if len(args) < 2 {
		o.fail("mv: missing file operand")
		return
	}
	o.fail("mv: cannot move '" + args[0] + "' to '" + args[len(args)-1] + "': " + problemReadOnly)
}
