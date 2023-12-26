package main

import (
	"syscall/js"
)

func typewriter(this js.Value, args []js.Value) interface{} {
	id := args[0]
	text := args[1].String()
	delay := args[2]
	document := js.Global().Get("document")
	element := document.Call("getElementById", id)
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

func handleAllButtonClicks(this js.Value, args []js.Value) interface{} {
	eventListener := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		event := p[0]
		target := event.Get("target")
		id := target.Get("id")
		tagName := target.Get("tagName")
		if tagName.String() == "BUTTON" {
			// TODO: Outgoing request to Slack
			js.Global().Get("console").Call("log", id)
		}

		return nil
	})

	document := js.Global().Get("document")
	body := document.Get("body")
	body.Call("addEventListener", "click", eventListener)

	return nil
}

func handleAllAnchorClicks(this js.Value, args []js.Value) interface{} {
	eventListener := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		event := p[0]
		target := event.Get("target")
		id := target.Get("id")
		tagName := target.Get("tagName")
		if tagName.String() == "A" {
			// TODO: Outgoing request to Slack
			js.Global().Get("console").Call("log", id)
		}

		return nil
	})

	document := js.Global().Get("document")
	body := document.Get("body")
	body.Call("addEventListener", "click", eventListener)

	return nil
}

func main() {
	c := make(chan struct{})

	js.Global().Set("typewriter", js.FuncOf(typewriter))
	js.Global().Set("handleAllButtonClicks", js.FuncOf(handleAllButtonClicks))
	js.Global().Set("handleAllAnchorClicks", js.FuncOf(handleAllAnchorClicks))

	<-c
}
