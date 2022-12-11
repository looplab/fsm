//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
)

func main() {
	fsm := fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "produce", Src: []string{"idle"}, Dst: "idle"},
			{Name: "consume", Src: []string{"idle"}, Dst: "idle"},
			{Name: "remove", Src: []string{"idle"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"produce": func(_ context.Context, e *fsm.Event) {
				e.FSM.SetMetadata("message", "hii")
				fmt.Println("produced data")
			},
			"consume": func(_ context.Context, e *fsm.Event) {
				message, ok := e.FSM.Metadata("message")
				if ok {
					fmt.Println("message = " + message.(string))
				}
			},
			"remove": func(_ context.Context, e *fsm.Event) {
				e.FSM.DeleteMetadata("message")
				if _, ok := e.FSM.Metadata("message"); !ok {
					fmt.Println("message removed")
				}
			},
		},
	)

	fmt.Println(fsm.Current())

	err := fsm.Event(context.Background(), "produce")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())

	err = fsm.Event(context.Background(), "consume")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())

	err = fsm.Event(context.Background(), "remove")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(fsm.Current())

}
