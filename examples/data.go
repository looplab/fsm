// +build ignore

package main

import (
	"fmt"

	"github.com/looplab/fsm"
)

func main() {
	fsm := fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "produce", Src: []string{"idle"}, Dst: "idle"},
			{Name: "consume", Src: []string{"idle"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"produce": func(e *fsm.Event) {
				e.FSM.SetMetadata("message", "hii")
				fmt.Println("produced data")
			},
			"consume": func(e *fsm.Event) {
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
