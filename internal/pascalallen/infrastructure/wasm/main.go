//go:build js && wasm

package main

import (
	"syscall/js"
	"time"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/shell"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/wasm/plain"
)

const maxDepth = 4

var functions = []string{"boot", "exec", "complete", "prompt", "history", "commands", "deliver", "deliverError"}

func fromJS(v js.Value, depth int) any {
	switch v.Type() {
	case js.TypeString:
		return v.String()
	case js.TypeNumber:
		return v.Float()
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeObject:
		if depth >= maxDepth {
			return nil
		}
		if js.Global().Get("Array").Call("isArray", v).Bool() {
			list := make([]any, v.Length())
			for i := range list {
				list[i] = fromJS(v.Index(i), depth+1)
			}
			return list
		}
		keys := js.Global().Get("Object").Call("keys", v)
		m := make(map[string]any, keys.Length())
		for i := 0; i < keys.Length(); i++ {
			key := keys.Index(i).String()
			m[key] = fromJS(v.Get(key), depth+1)
		}
		return m
	}

	return nil
}

func register(bridge *plain.Bridge) js.Value {
	object := js.Global().Get("Object").New()
	for _, name := range functions {
		object.Set(name, js.FuncOf(func(_ js.Value, args []js.Value) any {
			converted := make([]any, len(args))
			for i, a := range args {
				converted[i] = fromJS(a, 0)
			}
			return js.ValueOf(bridge.Call(name, converted))
		}))
	}

	return object
}

func main() {
	bridge := plain.New(shell.New(content.FS(), shell.WithClock(time.Now)))
	js.Global().Set("__shell", register(bridge))

	if ready := js.Global().Get("__shellReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}

	select {}
}
