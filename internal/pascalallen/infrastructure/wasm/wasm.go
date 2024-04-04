//go:build js && wasm

package main

import "syscall/js"

func typewriter(this js.Value, args []js.Value) interface{} {
	id := args[0]
	text := args[1].String()
	delay := args[2]
	document := js.Global().Get("document")
	element := document.Call("getElementById", id)

	if element.IsNull() {
		return nil
	}

	i := 0
	result := ""

	var f func()
	f = func() {
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, p []js.Value) interface{} {
			result += string(text[i])
			element.Set("textContent", result)
			i++
			if result != text {
				f()
			}

			return nil
		}), delay)
	}

	f()

	return nil
}

func main() {
	c := make(chan struct{}, 3)

	js.Global().Set("typewriter", js.FuncOf(typewriter))

	<-c
}
