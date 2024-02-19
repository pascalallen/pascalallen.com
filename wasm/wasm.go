//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall/js"
)

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

func handleAllButtonClicks(this js.Value, args []js.Value) interface{} {
	if os.Getenv("APP_ENV") != "prod" || os.Getenv("APP_ENV") != "production" {
		return nil
	}

	eventListener := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		event := p[0]
		target := event.Get("target")
		id := target.Get("id")
		tagName := target.Get("tagName")
		if tagName.String() == "BUTTON" {
			jsonBody := []byte(`{"text": ` + id.String() + `}`)
			bodyReader := bytes.NewReader(jsonBody)
			res, err := http.Post(os.Getenv("SLACK_DM_URL"), "application/json", bodyReader)
			if err != nil {
				js.Global().Get("console").Call("log", err)
				return nil
			}
			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode > 299 {
				js.Global().Get("console").Call("log", fmt.Sprintf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body))
				return nil
			}
			if err != nil {
				js.Global().Get("console").Call("log", err)
				return nil
			}
			js.Global().Get("console").Call("log", body)
		}

		return nil
	})

	document := js.Global().Get("document")
	body := document.Get("body")
	body.Call("addEventListener", "click", eventListener)

	return nil
}

func handleAllAnchorClicks(this js.Value, args []js.Value) interface{} {
	if os.Getenv("APP_ENV") != "prod" || os.Getenv("APP_ENV") != "production" {
		return nil
	}

	eventListener := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		event := p[0]
		target := event.Get("target")
		id := target.Get("id")
		tagName := target.Get("tagName")
		if tagName.String() == "A" {
			jsonBody := []byte(`{"text": ` + id.String() + `}`)
			bodyReader := bytes.NewReader(jsonBody)
			res, err := http.Post(os.Getenv("SLACK_DM_URL"), "application/json", bodyReader)
			if err != nil {
				js.Global().Get("console").Call("log", err)
				return nil
			}
			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode > 299 {
				js.Global().Get("console").Call("log", fmt.Sprintf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body))
				return nil
			}
			if err != nil {
				js.Global().Get("console").Call("log", err)
				return nil
			}
			js.Global().Get("console").Call("log", body)
		}

		return nil
	})

	document := js.Global().Get("document")
	body := document.Get("body")
	body.Call("addEventListener", "click", eventListener)

	return nil
}

func main() {
	c := make(chan struct{}, 3)

	js.Global().Set("typewriter", js.FuncOf(typewriter))

	<-c
}
