package shell

import (
	"strings"
	"time"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
)

var fixedNow = time.Date(2026, time.November, 11, 9, 30, 0, 0, time.UTC)

func newTestShell(opts ...Option) *Shell {
	return New(content.FS(), append([]Option{WithClock(func() time.Time { return fixedNow })}, opts...)...)
}

func lineText(l Line) string {
	var b strings.Builder
	for _, sp := range l {
		b.WriteString(sp.Text)
	}
	return b.String()
}

func resultText(r Result) string {
	var lines []string
	for _, b := range r.Blocks {
		for _, l := range b.Lines {
			lines = append(lines, lineText(l))
		}
	}
	return strings.Join(lines, "\n")
}

func errorText(r Result) string {
	var lines []string
	for _, b := range r.Blocks {
		if b.Kind != BlockError {
			continue
		}
		for _, l := range b.Lines {
			lines = append(lines, lineText(l))
		}
	}
	return strings.Join(lines, "\n")
}

func spans(r Result) []Span {
	var out []Span
	for _, b := range r.Blocks {
		for _, l := range b.Lines {
			out = append(out, l...)
		}
	}
	return out
}

func runs(r Result) []string {
	var out []string
	for _, sp := range spans(r) {
		if sp.Run != "" {
			out = append(out, sp.Run)
		}
	}
	return out
}

func hrefs(r Result) []string {
	var out []string
	for _, sp := range spans(r) {
		if sp.Href != "" {
			out = append(out, sp.Href)
		}
	}
	return out
}
