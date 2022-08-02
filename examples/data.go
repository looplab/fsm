//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"

	"github.com/sjqzhang/fsm"
)

func main() {
	fsm := fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "produce", Src: []string{"idle"}, Dst: "idle"},
			{Name: "consume", Src: []string{"idle"}, Dst: "idle"},
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

}
