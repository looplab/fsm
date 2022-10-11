//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
)

func main() {
	var afterFinishCalled bool
	fsm := fsm.NewFSM(
		"start",
		fsm.Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
			{Name: "finish", Src: []string{"end"}, Dst: "finished"},
			{Name: "reset", Src: []string{"end", "finished"}, Dst: "start"},
		},
		fsm.Callbacks{
			"enter_end": func(ctx context.Context, e *fsm.Event) {
				if err := e.FSM.Event(ctx, "finish"); err != nil {
					fmt.Println(err)
				}
			},
			"after_finish": func(ctx context.Context, e *fsm.Event) {
				afterFinishCalled = true
				if e.Src != "end" {
					panic(fmt.Sprintf("source should have been 'end' but was '%s'", e.Src))
				}
				if err := e.FSM.Event(ctx, "reset"); err != nil {
					fmt.Println(err)
				}
			},
		},
	)

	if err := fsm.Event(context.Background(), "run"); err != nil {
		panic(fmt.Sprintf("Error encountered when triggering the run event: %v", err))
	}

	if !afterFinishCalled {
		panic(fmt.Sprintf("After finish callback should have run, current state: '%s'", fsm.Current()))
	}

	currentState := fsm.Current()
	if currentState != "start" {
		panic(fmt.Sprintf("expected state to be 'start', was '%s'", currentState))
	}

	fmt.Println("Successfully ran state machine.")
}
