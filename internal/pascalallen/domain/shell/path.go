package shell

import (
	"io/fs"
	"strings"
)

const (
	problemMissing  = "No such file or directory"
	problemIsDir    = "Is a directory"
	problemNotDir   = "Not a directory"
	problemReadOnly = "Read-only file system"
)

func (s *Shell) home() string {
	return "/home/" + s.user
}

func (s *Shell) resolve(arg string) string {
	var segs []string
	switch {
	case arg == "~" || strings.HasPrefix(arg, "~/"):
		arg = arg[1:]
	case arg == s.home() || strings.HasPrefix(arg, s.home()+"/"):
		arg = arg[len(s.home()):]
	case strings.HasPrefix(arg, "/"):
	default:
		if s.cwd != "." {
			segs = strings.Split(s.cwd, "/")
		}
	}
	for _, part := range strings.Split(arg, "/") {
		switch part {
		case "", ".":
		case "..":
			if len(segs) > 0 {
				segs = segs[:len(segs)-1]
			}
		default:
			segs = append(segs, part)
		}
	}
	if len(segs) == 0 {
		return "."
	}
	return strings.Join(segs, "/")
}

func (s *Shell) stat(p string) (fs.FileInfo, string) {
	info, err := fs.Stat(s.fsys, p)
	if err != nil {
		return nil, problemMissing
	}
	return info, ""
}

func (s *Shell) read(arg string) (string, string) {
	p := s.resolve(arg)
	info, problem := s.stat(p)
	if problem != "" {
		return "", problem
	}
	if info.IsDir() {
		return "", problemIsDir
	}
	data, err := fs.ReadFile(s.fsys, p)
	if err != nil {
		return "", problemMissing
	}
	return string(data), ""
}

func (s *Shell) displayPath(p string) string {
	if p == "." {
		return s.home()
	}
	return s.home() + "/" + p
}

func tildePath(p string) string {
	if p == "." {
		return "~"
	}
	return "~/" + p
}

func shellQuote(word string) string {
	for _, r := range word {
		safe := r == '.' || r == '_' || r == '-' || r == '/' || r == '~' || r == '@' || r == '+' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		if !safe {
			return "'" + strings.ReplaceAll(word, "'", `'\''`) + "'"
		}
	}
	return word
}

func entrySpan(p string, name string, dir bool) Span {
	target := shellQuote(tildePath(p))
	if dir {
		return Span{Text: name + "/", Style: StylePath, Run: "ls " + target}
	}
	return Span{Text: name, Style: StylePlain, Run: "cat " + target}
}

func join(dir, name string) string {
	if dir == "." {
		return name
	}
	return dir + "/" + name
}
