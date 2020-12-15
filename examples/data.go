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
				e.FSM.WriteData("message", "hii")
				fmt.Println("produced data")
			},
			"consume": func(e *fsm.Event) {
				message := e.FSM.ReadData("message").(string)
				fmt.Println("message = " + message)
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
