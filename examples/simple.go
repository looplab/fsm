//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/looplab/fsm"
)

func main() {
	fsm, err := fsm.New(
		"closed",
		fsm.Transistions[string, string]{
			{Event: "open", Src: []string{"closed"}, Dst: "open"},
			{Event: "close", Src: []string{"open"}, Dst: "closed"},
		},
		fsm.Callbacks[string, string]{},
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())

	err = fsm.Event("open")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())

	err = fsm.Event("close")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())
}
