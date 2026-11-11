package render

import (
	"html"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

func Blocks(blocks []shell.Block) template.HTML {
	var b strings.Builder
	for _, block := range blocks {
		writeBlock(&b, block)
	}
	return template.HTML(b.String())
}

func BootEntry(result shell.Result) template.HTML {
	var b strings.Builder
	b.WriteString(`<div class="entry entry-boot" data-prerendered="">`)
	for _, block := range result.Blocks {
		writeBlock(&b, block)
	}
	b.WriteString(`</div>`)
	return template.HTML(b.String())
}

func writeBlock(b *strings.Builder, block shell.Block) {
	kind := "text"
	switch block.Kind {
	case shell.BlockError:
		kind = "error"
	case shell.BlockMedia:
		kind = "media"
	}
	b.WriteString(`<div class="block block-` + kind + `">`)
	if kind == "media" {
		if src, ok := safeMediaSrc(block.Src); ok {
			b.WriteString(`<img class="media" src="` + html.EscapeString(src) + `" alt="` + html.EscapeString(block.Alt) + `">`)
		}
	}
	for _, line := range block.Lines {
		class, spans := splitIndent(line)
		b.WriteString(`<div class="` + class + `">`)
		for _, span := range spans {
			writeSpan(b, span)
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
}

func isGap(text string) bool {
	return len(text) >= 2 && strings.Trim(text, " ") == ""
}

const maxIndent = 8

func splitIndent(line shell.Line) (string, shell.Line) {
	if len(line) == 0 {
		return "line", line
	}
	first := line[0].Text
	indent := len(first) - len(strings.TrimLeft(first, " "))
	if indent > 0 {
		var rest shell.Line
		if indent == len(first) {
			rest = line[1:]
		} else {
			rest = append(shell.Line{{Text: first[indent:], Style: line[0].Style, Href: line[0].Href, Run: line[0].Run}}, line[1:]...)
		}
		return "line indent-" + strconv.Itoa(min(indent, maxIndent)), rest
	}
	for _, span := range line {
		if span.Run != "" || isGap(span.Text) {
			return "line", line
		}
	}

	return "line line-prose", line
}

func writeSpan(b *strings.Builder, span shell.Span) {
	class := styleClass(span)
	text := html.EscapeString(span.Text)
	if span.Href != "" {
		if href, ok := safeHref(span.Href); ok {
			b.WriteString(`<a class="` + class + `" href="` + html.EscapeString(href) + `" target="_blank" rel="noopener noreferrer">` + text + `</a>`)
			return
		}
	}
	if span.Href == "" && span.Run != "" {
		b.WriteString(`<a class="run ` + class + `" href="/#` + encodeURIComponent(span.Run) + `" data-run="` + html.EscapeString(span.Run) + `">` + text + `</a>`)
		return
	}
	b.WriteString(`<span class="` + class + `">` + text + `</span>`)
}

func styleClass(span shell.Span) string {
	name := span.Style.String()
	if name == "unknown" {
		name = "plain"
	}
	class := "s-" + name
	if isGap(span.Text) {
		class += " gap"
	}
	return class
}

func safeHref(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	switch u.Scheme {
	case "http", "https":
		if u.Host == "" {
			return "", false
		}
		if u.Path == "" {
			u.Path = "/"
		}
		return u.String(), true
	case "mailto":
		if u.Opaque == "" {
			return "", false
		}
		return u.String(), true
	}
	return "", false
}

func safeMediaSrc(raw string) (string, bool) {
	if len(raw) < 2 || raw[0] != '/' || raw[1] == '/' || raw[1] == '\\' {
		return "", false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host != "" {
		return "", false
	}
	out := u.EscapedPath()
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	return out, true
}

func encodeURIComponent(s string) string {
	const unreserved = "-_.!~*'()"
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte(unreserved, c) >= 0 {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(hex[c>>4])
		b.WriteByte(hex[c&15])
	}
	return b.String()
}
