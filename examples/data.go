//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/looplab/fsm"
)

func main() {
	fsm := fsm.NewFSM(
		"idle",
		fsm.StateMachine[string, string]{
			{Event: "produce", Src: []string{"idle"}, Dst: "idle"},
			{Event: "consume", Src: []string{"idle"}, Dst: "idle"},
		},
		fsm.Callbacks[string, string]{
			"produce": func(e *fsm.CallbackReference[string, string]) {
				e.FSM.SetMetadata("message", "hii")
				fmt.Println("produced data")
			},
			"consume": func(e *fsm.CallbackReference[string, string]) {
				message, ok := e.FSM.Metadata("message")
				if ok {
					fmt.Println("message = " + message.(string))
				}

			},
		},
	)

	fmt.Println(fsm.Current())

	err := fsm.Event("produce")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())

	err = fsm.Event("consume")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())

}
