package main

import (
	"fmt"
	"syscall/js"
)

func echo(this js.Value, args []js.Value) interface{} {
	fmt.Println(args)

	return nil
}

func main() {
	c := make(chan struct{})

	js.Global().Set("echo", js.FuncOf(echo))

	<-c
}
