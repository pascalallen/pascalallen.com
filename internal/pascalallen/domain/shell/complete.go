package shell

import (
	"io/fs"
	"sort"
	"strings"
	"unicode/utf8"
)

type candidate struct {
	insert  string
	display string
	suffix  string
}

func (s *Shell) Complete(line string, cursor int) Completion {
	rs := []rune(line)
	cursor = max(0, min(cursor, len(rs)))
	if len(rs) > maxLineRunes {
		return Completion{Line: line, Cursor: cursor}
	}
	before, after := string(rs[:cursor]), string(rs[cursor:])
	start := strings.LastIndexAny(before, " \t;&") + 1
	head, word := before[:start], before[start:]
	unchanged := Completion{Line: line, Cursor: cursor}
	cands := s.candidates(head, word)
	if len(cands) == 0 {
		return unchanged
	}
	if len(cands) == 1 {
		insert := cands[0].insert
		if !strings.HasPrefix(after, cands[0].suffix) {
			insert += cands[0].suffix
		}
		return Completion{Line: head + insert + after, Cursor: utf8.RuneCountInString(head + insert)}
	}
	prefix := cands[0].insert
	displays := make([]string, 0, len(cands))
	for _, c := range cands {
		prefix = commonPrefix(prefix, c.insert)
		displays = append(displays, clean(c.display))
	}
	sort.Strings(displays)
	return Completion{Line: head + prefix + after, Cursor: utf8.RuneCountInString(head + prefix), Candidates: displays}
}

func (s *Shell) candidates(head, word string) []candidate {
	fields := strings.Fields(head[strings.LastIndexAny(head, ";&")+1:])
	if len(fields) == 0 {
		return s.commandCandidates(word)
	}
	c, _ := s.lookup(fields[0])
	switch c.args {
	case argPath:
		return s.pathCandidates(word, false)
	case argDir:
		return s.pathCandidates(word, true)
	case argCommand:
		return s.commandCandidates(word)
	case argTheme:
		var out []candidate
		for _, t := range themes {
			if strings.HasPrefix(t, word) {
				out = append(out, candidate{insert: t, display: t, suffix: " "})
			}
		}
		return out
	}
	return nil
}

func (s *Shell) commandCandidates(word string) []candidate {
	var out []candidate
	for _, c := range s.commands {
		if !c.hidden && strings.HasPrefix(c.name, word) {
			out = append(out, candidate{insert: c.name, display: c.name, suffix: " "})
		}
	}
	return out
}

func (s *Shell) pathCandidates(word string, dirsOnly bool) []candidate {
	if word == "~" {
		return []candidate{{insert: "~/", display: "~/"}}
	}
	slash := strings.LastIndexByte(word, '/')
	dirPart, base := word[:slash+1], word[slash+1:]
	dir, _ := s.resolve(dirPart)
	entries, _ := fs.ReadDir(s.fsys, dir)
	var out []candidate
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, base) || (dirsOnly && !e.IsDir()) {
			continue
		}
		if e.IsDir() {
			out = append(out, candidate{insert: dirPart + name + "/", display: name + "/"})
		} else {
			out = append(out, candidate{insert: dirPart + name, display: name, suffix: " "})
		}
	}
	return out
}

func commonPrefix(a, b string) string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	for i > 0 && i < len(a) && !utf8.RuneStart(a[i]) {
		i--
	}
	return a[:i]
}
