//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/looplab/fsm/v2"
)

func main() {
	fsm, err := fsm.New(
		"idle",
		fsm.Transitions[string, string]{
			{Event: "produce", Src: []string{"idle"}, Dst: "idle"},
			{Event: "consume", Src: []string{"idle"}, Dst: "idle"},
		},
		fsm.Callbacks[string, string]{
			fsm.Callback[string, string]{When: fsm.BeforeEvent, Event: "sproduce",
				F: func(e *fsm.CallbackContext[string, string]) {
					e.FSM.SetMetadata("message", "hii")
					fmt.Println("produced data")
				},
			},
			fsm.Callback[string, string]{When: fsm.BeforeEvent, Event: "consume",
				F: func(e *fsm.CallbackContext[string, string]) {
					message, ok := e.FSM.Metadata("message")
					if ok {
						fmt.Println("message = " + message.(string))
					}
				},
			},
		},
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())

	err = fsm.Event("produce")
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
