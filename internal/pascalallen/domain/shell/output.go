package shell

import (
	"strconv"
	"strings"
	"unicode"
)

type output struct {
	blocks    []Block
	effects   []Effect
	failed    bool
	count     int
	truncated bool
}

func (o *output) add(kind BlockKind, line Line) {
	if o.count >= maxOutputLines {
		o.truncated = true
		return
	}
	o.count++
	if n := len(o.blocks); n > 0 && o.blocks[n-1].Kind == kind {
		o.blocks[n-1].Lines = append(o.blocks[n-1].Lines, line)
		return
	}
	o.blocks = append(o.blocks, Block{Kind: kind, Lines: []Line{line}})
}

func (o *output) line(spans ...Span) {
	o.add(BlockText, Line(append([]Span{}, spans...)))
}

func (o *output) lines(lines []Line) {
	for _, l := range lines {
		o.line(l...)
	}
}

func (o *output) fail(msg string) {
	o.failLine(Span{Text: msg, Style: StyleError})
}

func (o *output) failLine(spans ...Span) {
	o.failed = true
	o.add(BlockError, Line(append([]Span{}, spans...)))
}

func (o *output) media(src, alt string, caption Line) {
	o.blocks = append(o.blocks, Block{Kind: BlockMedia, Src: src, Alt: alt, Lines: []Line{caption}})
}

func (o *output) effect(kind EffectKind, arg string) {
	o.effects = append(o.effects, Effect{Kind: kind, Arg: arg})
}

func (o *output) clear() {
	o.blocks = nil
	o.count = 0
	o.truncated = false
	o.effect(EffectClear, "")
}

func (o *output) result() Result {
	if o.truncated {
		o.count, o.truncated = 0, false
		o.add(BlockText, Line{styled("output truncated after "+strconv.Itoa(maxOutputLines)+" lines", StyleDim)})
	}
	r := Result{Blocks: o.blocks, Effects: o.effects}
	for bi := range r.Blocks {
		b := &r.Blocks[bi]
		b.Src = clean(b.Src)
		b.Alt = clean(b.Alt)
		for li := range b.Lines {
			for si := range b.Lines[li] {
				sp := &b.Lines[li][si]
				sp.Text = clean(sp.Text)
				sp.Href = clean(sp.Href)
				sp.Run = clean(sp.Run)
			}
		}
	}
	for i := range r.Effects {
		r.Effects[i].Arg = clean(r.Effects[i].Arg)
	}
	return r
}

func clean(s string) string {
	s = strings.ToValidUTF8(s, "\uFFFD")
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.Is(unicode.Cf, r):
			return -1
		case unicode.IsControl(r) || unicode.In(r, unicode.Zl, unicode.Zp):
			return ' '
		}
		return r
	}, s)
}

func plain(text string) Span { return Span{Text: text} }

func styled(text string, style Style) Span { return Span{Text: text, Style: style} }

func runSpan(text, command string) Span { return Span{Text: text, Style: StyleAccent, Run: command} }

func linkSpan(text, href string) Span { return Span{Text: text, Style: StyleAccent, Href: href} }
