//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/looplab/fsm"
)

func main() {
	f := fsm.NewFSM(
		"start",
		fsm.Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		fsm.Callbacks{
			"leave_start": func(_ context.Context, e *fsm.Event) {
				e.Async()
			},
		},
	)

	err := f.Event(context.Background(), "run")
	asyncError, ok := err.(fsm.AsyncError)
	if !ok {
		panic(fmt.Sprintf("expected error to be 'AsyncError', got %v", err))
	}
	var asyncStateTransitionWasCanceled bool
	go func() {
		<-asyncError.Ctx.Done()
		asyncStateTransitionWasCanceled = true
		if asyncError.Ctx.Err() != context.Canceled {
			panic(fmt.Sprintf("Expected error to be '%v' but was '%v'", context.Canceled, asyncError.Ctx.Err()))
		}
	}()
	asyncError.CancelTransition()
	time.Sleep(20 * time.Millisecond)

	if err = f.Transition(); err != nil {
		panic(fmt.Sprintf("Error encountered when transitioning: %v", err))
	}
	if !asyncStateTransitionWasCanceled {
		panic("expected async state transition cancelation to have propagated")
	}
	if f.Current() != "start" {
		panic("expected state to be 'start'")
	}

	fmt.Println("Successfully ran state machine.")
}
