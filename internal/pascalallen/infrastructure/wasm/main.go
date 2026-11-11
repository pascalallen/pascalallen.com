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

var functions = []string{"boot", "exec", "complete", "prompt", "history", "commands", "deliver", "deliverError", "setTheme"}

func global(name string) js.Value {
	return js.Global().Get(name)
}

func ownData(v js.Value, key js.Value) (js.Value, bool) {
	descriptor := global("Object").Call("getOwnPropertyDescriptor", v, key)
	if descriptor.Type() != js.TypeObject || !descriptor.Call("hasOwnProperty", "value").Bool() {
		return js.Undefined(), false
	}

	return descriptor.Get("value"), true
}

func isPlainObject(v js.Value) bool {
	prototype := global("Object").Call("getPrototypeOf", v)
	return prototype.IsNull() || prototype.Equal(global("Object").Get("prototype"))
}

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
		if global("Array").Call("isArray", v).Bool() {
			list := make([]any, v.Length())
			for i := range list {
				element, ok := ownData(v, js.ValueOf(i))
				if !ok {
					return nil
				}
				list[i] = fromJS(element, depth+1)
			}
			return list
		}
		if !isPlainObject(v) {
			return nil
		}
		keys := global("Object").Call("keys", v)
		m := make(map[string]any, keys.Length())
		for i := 0; i < keys.Length(); i++ {
			key := keys.Index(i)
			value, ok := ownData(v, key)
			if !ok {
				return nil
			}
			m[key.String()] = fromJS(value, depth+1)
		}
		return m
	}

	return nil
}

func register(bridge *plain.Bridge) js.Value {
	object := js.Global().Get("Object").New()
	for _, name := range functions {
		object.Set(name, js.FuncOf(func(_ js.Value, args []js.Value) any {
			return js.ValueOf(bridge.CallLazy(name, func() []any {
				converted := make([]any, len(args))
				for i, a := range args {
					converted[i] = fromJS(a, 0)
				}
				return converted
			}))
		}))
	}

	return object
}

func notifyReady() {
	defer func() {
		if recover() != nil {
			js.Global().Get("console").Call("error", "shell: __shellReady threw")
		}
	}()
	if ready := js.Global().Get("__shellReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}
}

func main() {
	bridge := plain.New(shell.New(content.FS(), shell.WithClock(time.Now)))
	js.Global().Set("__shell", register(bridge))

	notifyReady()

	select {}
}
