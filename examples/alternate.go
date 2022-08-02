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
			{Name: "scan", Src: []string{"idle"}, Dst: "scanning"},
			{Name: "working", Src: []string{"scanning"}, Dst: "scanning"},
			{Name: "situation", Src: []string{"scanning"}, Dst: "scanning"},
			{Name: "situation", Src: []string{"idle"}, Dst: "idle"},
			{Name: "finish", Src: []string{"scanning"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"scan": func(_ context.Context, e *fsm.Event) {
				fmt.Println("after_scan: " + e.FSM.Current())
			},
			"working": func(_ context.Context, e *fsm.Event) {
				fmt.Println("working: " + e.FSM.Current())
			},
			"situation": func(_ context.Context, e *fsm.Event) {
				fmt.Println("situation: " + e.FSM.Current())
			},
			"finish": func(_ context.Context, e *fsm.Event) {
				fmt.Println("finish: " + e.FSM.Current())
			},
		},
	)

	fmt.Println(fsm.Current())

	err := fsm.Event(context.Background(), "scan")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("1:" + fsm.Current())

	err = fsm.Event(context.Background(), "working")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("2:" + fsm.Current())

	err = fsm.Event(context.Background(), "situation")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("3:" + fsm.Current())

	err = fsm.Event(context.Background(), "finish")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("4:" + fsm.Current())

}
