package shell

import "strings"

func renderText(text string) []Line {
	text = strings.TrimSuffix(text, "\n")
	if text == "" {
		return nil
	}
	raw := strings.Split(text, "\n")
	lines := make([]Line, 0, len(raw))
	for _, l := range raw {
		if strings.HasPrefix(l, "# ") {
			lines = append(lines, Line{styled(l[2:], StyleHeading)})
			continue
		}
		lines = append(lines, linkify(l, StylePlain))
	}
	return lines
}

func linkify(text string, style Style) Line {
	line := Line{}
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			line = append(line, styled(buf.String(), style))
			buf.Reset()
		}
	}
	for i := 0; i < len(text); {
		if text[i] == ' ' {
			buf.WriteByte(' ')
			i++
			continue
		}
		j := strings.IndexByte(text[i:], ' ')
		if j < 0 {
			j = len(text)
		} else {
			j += i
		}
		token := text[i:j]
		i = j
		lead, core, trail := trimPunct(token)
		href := linkTarget(core)
		if href == "" {
			buf.WriteString(token)
			continue
		}
		buf.WriteString(lead)
		flush()
		line = append(line, linkSpan(core, href))
		buf.WriteString(trail)
	}
	flush()
	return line
}

func trimPunct(token string) (string, string, string) {
	start, end := 0, len(token)
	for start < end && strings.IndexByte("(<[\"'", token[start]) >= 0 {
		start++
	}
	for end > start && strings.IndexByte(".,;:!?)>]\"'", token[end-1]) >= 0 {
		end--
	}
	return token[:start], token[start:end], token[end:]
}

func linkTarget(word string) string {
	if isWebURL(word) {
		return word
	}
	if isEmail(word) {
		return "mailto:" + word
	}
	return ""
}

func isWebURL(word string) bool {
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(word, scheme) && len(word) > len(scheme) {
			return true
		}
	}
	return false
}

func isMailto(word string) bool {
	if !strings.HasPrefix(word, "mailto:") {
		return false
	}
	addr := word[len("mailto:"):]
	if i := strings.IndexByte(addr, '?'); i >= 0 {
		addr = addr[:i]
	}
	return isEmail(addr)
}

func isEmail(word string) bool {
	at := strings.IndexByte(word, '@')
	if at <= 0 || strings.Count(word, "@") != 1 {
		return false
	}
	domain := word[at+1:]
	dot := strings.LastIndexByte(domain, '.')
	if dot <= 0 || dot == len(domain)-1 {
		return false
	}
	for _, r := range word {
		ok := r == '.' || r == '_' || r == '-' || r == '+' || r == '@' ||
			(r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		if !ok {
			return false
		}
	}
	return true
}

func firstURL(lines []Line) string {
	for _, l := range lines {
		for _, sp := range l {
			if isWebURL(sp.Href) {
				return sp.Href
			}
		}
	}
	return ""
}

func summarize(text string) (string, string) {
	lines := strings.Split(text, "\n")
	var title []string
	i := 0
	for ; i < len(lines) && strings.HasPrefix(lines[i], "# "); i++ {
		title = append(title, lines[i][2:])
	}
	for ; i < len(lines); i++ {
		if t := strings.TrimSpace(lines[i]); t != "" && !strings.HasPrefix(lines[i], "# ") {
			return strings.Join(title, " "), t
		}
	}
	return strings.Join(title, " "), ""
}
