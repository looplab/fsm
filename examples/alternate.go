//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/looplab/fsm/v2"
)

func main() {
	f, err := fsm.New(
		"idle",
		fsm.Transitions[string, string]{
			{Event: "scan", Src: []string{"idle"}, Dst: "scanning"},
			{Event: "working", Src: []string{"scanning"}, Dst: "scanning"},
			{Event: "situation", Src: []string{"scanning"}, Dst: "scanning"},
			{Event: "situation", Src: []string{"idle"}, Dst: "idle"},
			{Event: "finish", Src: []string{"scanning"}, Dst: "idle"},
		},
		fsm.Callbacks[string, string]{
			fsm.Callback[string, string]{When: fsm.BeforeEvent, Event: "scan",
				F: func(e *fsm.CallbackContext[string, string]) {
					fmt.Println("after_scan: " + e.FSM.Current())
				},
			},
			fsm.Callback[string, string]{When: fsm.BeforeEvent, Event: "working",
				F: func(e *fsm.CallbackContext[string, string]) {
					fmt.Println("working: " + e.FSM.Current())
				},
			},
			fsm.Callback[string, string]{When: fsm.BeforeEvent, Event: "situation",
				F: func(e *fsm.CallbackContext[string, string]) {
					fmt.Println("situation: " + e.FSM.Current())
				},
			},
			fsm.Callback[string, string]{When: fsm.BeforeEvent, Event: "finish",
				F: func(e *fsm.CallbackContext[string, string]) {
					fmt.Println("finish: " + e.FSM.Current())
				},
			},
		},
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(f.Current())

	err = f.Event("scan")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("1:" + f.Current())

	err = f.Event("working")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("2:" + f.Current())

	err = f.Event("situation")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("3:" + f.Current())

	err = f.Event("finish")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("4:" + f.Current())

}
