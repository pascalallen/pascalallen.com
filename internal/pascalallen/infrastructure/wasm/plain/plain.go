package plain

import (
	"fmt"
	"math"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
)

type Bridge struct {
	sh *shell.Shell
}

func New(sh *shell.Shell) *Bridge {
	return &Bridge{sh: sh}
}

func Result(r shell.Result) map[string]any {
	blocks := make([]any, 0, len(r.Blocks))
	for _, b := range r.Blocks {
		lines := make([]any, 0, len(b.Lines))
		for _, l := range b.Lines {
			spans := make([]any, 0, len(l))
			for _, s := range l {
				spans = append(spans, map[string]any{
					"text":  s.Text,
					"style": s.Style.String(),
					"href":  s.Href,
					"run":   s.Run,
				})
			}
			lines = append(lines, spans)
		}
		blocks = append(blocks, map[string]any{
			"kind":  b.Kind.String(),
			"lines": lines,
			"src":   b.Src,
			"alt":   b.Alt,
		})
	}
	effects := make([]any, 0, len(r.Effects))
	for _, e := range r.Effects {
		effects = append(effects, map[string]any{"kind": e.Kind.String(), "arg": e.Arg})
	}

	return map[string]any{"blocks": blocks, "effects": effects}
}

func Completion(c shell.Completion) map[string]any {
	return map[string]any{
		"line":       c.Line,
		"cursor":     c.Cursor,
		"candidates": stringList(c.Candidates),
	}
}

func stringList(in []string) []any {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}

	return out
}

func failure(message string) map[string]any {
	return Result(shell.Result{Blocks: []shell.Block{{
		Kind:  shell.BlockError,
		Lines: []shell.Line{{{Text: message, Style: shell.StyleError}}},
	}}})
}

func (b *Bridge) Call(name string, args []any) any {
	return b.CallLazy(name, func() []any { return args })
}

func (b *Bridge) CallLazy(name string, convert func() []any) (out any) {
	defer func() {
		if r := recover(); r != nil {
			out = failure(fmt.Sprintf("shell: %s failed", name))
		}
	}()

	args := convert()

	switch name {
	case "boot":
		if len(args) != 0 {
			return failure("shell: boot expects no arguments")
		}
		return Result(b.sh.Boot())
	case "exec":
		line, ok := stringArg(args, 0, 1)
		if !ok {
			return failure("shell: exec expects 1 string argument")
		}
		return Result(b.sh.Exec(line))
	case "complete":
		line, ok := stringArg(args, 0, 2)
		if !ok {
			return failure("shell: complete expects a string and a number")
		}
		cursor, ok := args[1].(float64)
		if !ok || cursor != math.Trunc(cursor) || cursor < 0 || cursor > math.MaxInt32 {
			return failure("shell: complete expects a non-negative integer cursor")
		}
		return Completion(b.sh.Complete(line, int(cursor)))
	case "prompt":
		if len(args) != 0 {
			return failure("shell: prompt expects no arguments")
		}
		return b.sh.Prompt()
	case "history":
		if len(args) != 0 {
			return failure("shell: history expects no arguments")
		}
		return stringList(b.sh.History())
	case "commands":
		if len(args) != 0 {
			return failure("shell: commands expects no arguments")
		}
		infos := b.sh.Commands()
		list := make([]any, 0, len(infos))
		for _, c := range infos {
			list = append(list, map[string]any{"name": c.Name, "synopsis": c.Synopsis, "summary": c.Summary})
		}
		return list
	case "deliver":
		feed, ok := stringArg(args, 0, 2)
		if !ok {
			return failure("shell: deliver expects a feed name and an array of items")
		}
		items, ok := feedItems(args[1])
		if !ok {
			return failure("shell: deliver items must be objects with string name, description, url and meta")
		}
		return Result(b.sh.Deliver(feed, items))
	case "deliverError":
		feed, ok := stringArg(args, 0, 1)
		if !ok {
			return failure("shell: deliverError expects 1 string argument")
		}
		return Result(b.sh.DeliverError(feed))
	case "setTheme":
		theme, ok := stringArg(args, 0, 1)
		if !ok {
			return failure("shell: setTheme expects 1 string argument")
		}
		return b.sh.SetTheme(theme)
	}

	return failure("shell: unknown function " + name)
}

func stringArg(args []any, index, count int) (string, bool) {
	if len(args) != count {
		return "", false
	}
	s, ok := args[index].(string)

	return s, ok
}

func feedItems(v any) ([]shell.FeedItem, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	items := make([]shell.FeedItem, 0, len(list))
	for _, raw := range list {
		m, ok := raw.(map[string]any)
		if !ok {
			return nil, false
		}
		var item shell.FeedItem
		for key, dst := range map[string]*string{
			"name": &item.Name, "description": &item.Description, "url": &item.URL, "meta": &item.Meta,
		} {
			s, ok := m[key].(string)
			if !ok {
				return nil, false
			}
			*dst = s
		}
		items = append(items, item)
	}

	return items, true
}
