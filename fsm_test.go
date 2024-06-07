// Copyright (c) 2013 - Max Persson <max@looplab.se>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fsm

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

type fakeTransitionerObj struct {
}

func (t fakeTransitionerObj) transition(f *FSM) error {
	return &InternalError{}
}

func TestSameState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "start"},
		},
		Callbacks{},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestSetState(t *testing.T) {
	fsm := NewFSM(
		"walking",
		Events{
			{Name: "walk", Src: []string{"start"}, Dst: "walking"},
		},
		Callbacks{},
	)
	fsm.SetState("start")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'walking'")
	}
	err := fsm.Event(context.Background(), "walk")
	if err != nil {
		t.Error("transition is expected no error")
	}
}

func TestBadTransition(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "running"},
		},
		Callbacks{},
	)
	fsm.transitionerObj = new(fakeTransitionerObj)
	err := fsm.Event(context.Background(), "run")
	if err == nil {
		t.Error("bad transition should give an error")
	}
}

func TestInappropriateEvent(t *testing.T) {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	err := fsm.Event(context.Background(), "close")
	if e, ok := err.(InvalidEventError); !ok && e.Event != "close" && e.State != "closed" {
		t.Error("expected 'InvalidEventError' with correct state and event")
	}
}

func TestInvalidEvent(t *testing.T) {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	err := fsm.Event(context.Background(), "lock")
	if e, ok := err.(UnknownEventError); !ok && e.Event != "close" {
		t.Error("expected 'UnknownEventError' with correct event")
	}
}

func TestMultipleSources(t *testing.T) {
	fsm := NewFSM(
		"one",
		Events{
			{Name: "first", Src: []string{"one"}, Dst: "two"},
			{Name: "second", Src: []string{"two"}, Dst: "three"},
			{Name: "reset", Src: []string{"one", "two", "three"}, Dst: "one"},
		},
		Callbacks{},
	)

	err := fsm.Event(context.Background(), "first")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "two" {
		t.Error("expected state to be 'two'")
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "one" {
		t.Error("expected state to be 'one'")
	}
	err = fsm.Event(context.Background(), "first")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	err = fsm.Event(context.Background(), "second")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "three" {
		t.Error("expected state to be 'three'")
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "one" {
		t.Error("expected state to be 'one'")
	}
}

func TestMultipleEvents(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "first", Src: []string{"start"}, Dst: "one"},
			{Name: "second", Src: []string{"start"}, Dst: "two"},
			{Name: "reset", Src: []string{"one"}, Dst: "reset_one"},
			{Name: "reset", Src: []string{"two"}, Dst: "reset_two"},
			{Name: "reset", Src: []string{"reset_one", "reset_two"}, Dst: "start"},
		},
		Callbacks{},
	)

	err := fsm.Event(context.Background(), "first")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "reset_one" {
		t.Error("expected state to be 'reset_one'")
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}

	err = fsm.Event(context.Background(), "second")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "reset_two" {
		t.Error("expected state to be 'reset_two'")
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestGenericCallbacks(t *testing.T) {
	beforeEvent := false
	leaveState := false
	enterState := false
	afterEvent := false

	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_event": func(_ context.Context, e *Event) {
				beforeEvent = true
			},
			"leave_state": func(_ context.Context, e *Event) {
				leaveState = true
			},
			"enter_state": func(_ context.Context, e *Event) {
				enterState = true
			},
			"after_event": func(_ context.Context, e *Event) {
				afterEvent = true
			},
		},
	)

	err := fsm.Event(context.Background(), "run")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

func TestSpecificCallbacks(t *testing.T) {
	beforeEvent := false
	leaveState := false
	enterState := false
	afterEvent := false

	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_run": func(_ context.Context, e *Event) {
				beforeEvent = true
			},
			"leave_start": func(_ context.Context, e *Event) {
				leaveState = true
			},
			"enter_end": func(_ context.Context, e *Event) {
				enterState = true
			},
			"after_run": func(_ context.Context, e *Event) {
				afterEvent = true
			},
		},
	)

	err := fsm.Event(context.Background(), "run")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

func TestSpecificCallbacksShortform(t *testing.T) {
	enterState := false
	afterEvent := false

	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"end": func(_ context.Context, e *Event) {
				enterState = true
			},
			"run": func(_ context.Context, e *Event) {
				afterEvent = true
			},
		},
	)

	err := fsm.Event(context.Background(), "run")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if !(enterState && afterEvent) {
		t.Error("expected all callbacks to be called")
	}
}

func TestBeforeEventWithoutTransition(t *testing.T) {
	beforeEvent := true

	fsm := NewFSM(
		"start",
		Events{
			{Name: "dontrun", Src: []string{"start"}, Dst: "start"},
		},
		Callbacks{
			"before_event": func(_ context.Context, e *Event) {
				beforeEvent = true
			},
		},
	)

	err := fsm.Event(context.Background(), "dontrun")
	if e, ok := err.(NoTransitionError); !ok && e.Err != nil {
		t.Error("expected 'NoTransitionError' without custom error")
	}

	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
	if !beforeEvent {
		t.Error("expected callback to be called")
	}
}

func TestCancelBeforeGenericEvent(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_event": func(_ context.Context, e *Event) {
				e.Cancel()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelBeforeSpecificEvent(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_run": func(_ context.Context, e *Event) {
				e.Cancel()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelLeaveGenericState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_state": func(_ context.Context, e *Event) {
				e.Cancel()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelLeaveSpecificState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_start": func(_ context.Context, e *Event) {
				e.Cancel()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCancelWithError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_event": func(_ context.Context, e *Event) {
				e.Cancel(fmt.Errorf("error"))
			},
		},
	)
	err := fsm.Event(context.Background(), "run")
	if _, ok := err.(CanceledError); !ok {
		t.Error("expected only 'CanceledError'")
	}

	if e, ok := err.(CanceledError); ok && e.Err.Error() != "error" {
		t.Error("expected 'CanceledError' with correct custom error")
	}

	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestAsyncTransitionGenericState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_state": func(_ context.Context, e *Event) {
				e.Async()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
	err := fsm.Transition()
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "end" {
		t.Error("expected state to be 'end'")
	}
}

func TestAsyncTransitionSpecificState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_start": func(_ context.Context, e *Event) {
				e.Async()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
	err := fsm.Transition()
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "end" {
		t.Error("expected state to be 'end'")
	}
}

func TestAsyncTransitionInProgress(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
			{Name: "reset", Src: []string{"end"}, Dst: "start"},
		},
		Callbacks{
			"leave_start": func(_ context.Context, e *Event) {
				e.Async()
			},
		},
	)
	_ = fsm.Event(context.Background(), "run")
	err := fsm.Event(context.Background(), "reset")
	if e, ok := err.(InTransitionError); !ok && e.Event != "reset" {
		t.Error("expected 'InTransitionError' with correct state")
	}
	err = fsm.Transition()
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	err = fsm.Event(context.Background(), "reset")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestAsyncTransitionNotInProgress(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
			{Name: "reset", Src: []string{"end"}, Dst: "start"},
		},
		Callbacks{},
	)
	err := fsm.Transition()
	if _, ok := err.(NotInTransitionError); !ok {
		t.Error("expected 'NotInTransitionError'")
	}
}

func TestCancelAsyncTransition(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_start": func(_ context.Context, e *Event) {
				e.Async()
			},
		},
	)
	err := fsm.Event(context.Background(), "run")
	asyncError, ok := err.(AsyncError)
	if !ok {
		t.Errorf("expected error to be 'AsyncError', got %v", err)
	}
	var asyncStateTransitionWasCanceled = make(chan struct{})
	go func() {
		<-asyncError.Ctx.Done()
		close(asyncStateTransitionWasCanceled)
	}()
	asyncError.CancelTransition()
	<-asyncStateTransitionWasCanceled

	if err = fsm.Transition(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if fsm.Current() != "start" {
		t.Error("expected state to be 'start'")
	}
}

func TestCallbackNoError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(_ context.Context, e *Event) {
			},
		},
	)
	e := fsm.Event(context.Background(), "run")
	if e != nil {
		t.Error("expected no error")
	}
}

func TestCallbackError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(_ context.Context, e *Event) {
				e.Err = fmt.Errorf("error")
			},
		},
	)
	e := fsm.Event(context.Background(), "run")
	if e.Error() != "error" {
		t.Error("expected error to be 'error'")
	}
}

func TestCallbackArgs(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(_ context.Context, e *Event) {
				if len(e.Args) != 1 {
					t.Error("too few arguments")
				}
				arg, ok := e.Args[0].(string)
				if !ok {
					t.Error("not a string argument")
				}
				if arg != "test" {
					t.Error("incorrect argument")
				}
			},
		},
	)
	err := fsm.Event(context.Background(), "run", "test")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
}

func TestCallbackPanic(t *testing.T) {
	panicMsg := "unexpected panic"
	defer func() {
		r := recover()
		if r == nil || r != panicMsg {
			t.Errorf("expected panic message to be '%s', got %v", panicMsg, r)
		}
	}()
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(_ context.Context, e *Event) {
				panic(panicMsg)
			},
		},
	)
	e := fsm.Event(context.Background(), "run")
	if e.Error() != "error" {
		t.Error("expected error to be 'error'")
	}
}

func TestNoDeadLock(t *testing.T) {
	var fsm *FSM
	fsm = NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(_ context.Context, e *Event) {
				fsm.Current() // Should not result in a panic / deadlock
			},
		},
	)
	err := fsm.Event(context.Background(), "run")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
}

func TestThreadSafetyRaceCondition(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(_ context.Context, e *Event) {
			},
		},
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = fsm.Current()
	}()
	err := fsm.Event(context.Background(), "run")
	if err != nil {
		t.Errorf("transition failed %v", err)
	}
	wg.Wait()
}

func TestDoubleTransition(t *testing.T) {
	var fsm *FSM
	var wg sync.WaitGroup
	wg.Add(2)
	fsm = NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_run": func(_ context.Context, e *Event) {
				wg.Done()
				// Imagine a concurrent event coming in of the same type while
				// the data access mutex is unlocked because the current transition
				// is running its event callbacks, getting around the "active"
				// transition checks
				if len(e.Args) == 0 {
					// Must be concurrent so the test may pass when we add a mutex that synchronizes
					// calls to Event(...). It will then fail as an inappropriate transition as we
					// have changed state.
					go func() {
						if err := fsm.Event(context.Background(), "run", "second run"); err != nil {
							fmt.Println(err)
							wg.Done() // It should fail, and then we unfreeze the test.
						}
					}()
					time.Sleep(20 * time.Millisecond)
				} else {
					panic("Was able to reissue an event mid-transition")
				}
			},
		},
	)
	if err := fsm.Event(context.Background(), "run"); err != nil {
		fmt.Println(err)
	}
	wg.Wait()
}

func TestTransitionInCallbacks(t *testing.T) {
	var fsm *FSM
	var afterFinishCalled bool
	fsm = NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
			{Name: "finish", Src: []string{"end"}, Dst: "finished"},
			{Name: "reset", Src: []string{"end", "finished"}, Dst: "start"},
		},
		Callbacks{
			"enter_end": func(ctx context.Context, e *Event) {
				if err := e.FSM.Event(ctx, "finish"); err != nil {
					fmt.Println(err)
				}
			},
			"after_finish": func(ctx context.Context, e *Event) {
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
		t.Errorf("expected no error, got %v", err)
	}
	if !afterFinishCalled {
		t.Error("expected after_finish callback to have been executed but it wasn't")
	}

	currentState := fsm.Current()
	if currentState != "start" {
		t.Errorf("expected state to be 'start', was '%s'", currentState)
	}
}

func TestContextInCallbacks(t *testing.T) {
	var fsm *FSM
	var enterEndAsyncWorkDone = make(chan struct{})
	fsm = NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
			{Name: "finish", Src: []string{"end"}, Dst: "finished"},
			{Name: "reset", Src: []string{"end", "finished"}, Dst: "start"},
		},
		Callbacks{
			"enter_end": func(ctx context.Context, e *Event) {
				go func() {
					<-ctx.Done()
					close(enterEndAsyncWorkDone)
				}()

				<-ctx.Done()
				if err := e.FSM.Event(ctx, "finish"); err != nil {
					e.Err = fmt.Errorf("transitioning to the finished state failed: %w", err)
				}
			},
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
	}()
	err := fsm.Event(ctx, "run")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected 'context canceled' error, got %v", err)
	}
	<-enterEndAsyncWorkDone

	currentState := fsm.Current()
	if currentState != "end" {
		t.Errorf("expected state to be 'end', was '%s'", currentState)
	}
}

func TestNoTransition(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "start"},
		},
		Callbacks{},
	)
	err := fsm.Event(context.Background(), "run")
	if _, ok := err.(NoTransitionError); !ok {
		t.Error("expected 'NoTransitionError'")
	}
}

func TestNoTransitionAfterEventCallbackTransition(t *testing.T) {
	var fsm *FSM
	fsm = NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "start"},
			{Name: "finish", Src: []string{"start"}, Dst: "finished"},
		},
		Callbacks{
			"after_event": func(_ context.Context, e *Event) {
				fsm.Event(context.Background(), "finish")
			},
		},
	)
	err := fsm.Event(context.Background(), "run")
	if _, ok := err.(NoTransitionError); !ok {
		t.Error("expected 'NoTransitionError'")
	}

	currentState := fsm.Current()
	if currentState != "finished" {
		t.Errorf("expected state to be 'finished', was '%s'", currentState)
	}
}

func ExampleNewFSM() {
	fsm := NewFSM(
		"green",
		Events{
			{Name: "warn", Src: []string{"green"}, Dst: "yellow"},
			{Name: "panic", Src: []string{"yellow"}, Dst: "red"},
			{Name: "panic", Src: []string{"green"}, Dst: "red"},
			{Name: "calm", Src: []string{"red"}, Dst: "yellow"},
			{Name: "clear", Src: []string{"yellow"}, Dst: "green"},
		},
		Callbacks{
			"before_warn": func(_ context.Context, e *Event) {
				fmt.Println("before_warn")
			},
			"before_event": func(_ context.Context, e *Event) {
				fmt.Println("before_event")
			},
			"leave_green": func(_ context.Context, e *Event) {
				fmt.Println("leave_green")
			},
			"leave_state": func(_ context.Context, e *Event) {
				fmt.Println("leave_state")
			},
			"enter_yellow": func(_ context.Context, e *Event) {
				fmt.Println("enter_yellow")
			},
			"enter_state": func(_ context.Context, e *Event) {
				fmt.Println("enter_state")
			},
			"after_warn": func(_ context.Context, e *Event) {
				fmt.Println("after_warn")
			},
			"after_event": func(_ context.Context, e *Event) {
				fmt.Println("after_event")
			},
		},
	)
	fmt.Println(fsm.Current())
	err := fsm.Event(context.Background(), "warn")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// green
	// before_warn
	// before_event
	// leave_green
	// leave_state
	// enter_yellow
	// enter_state
	// after_warn
	// after_event
	// yellow
}

func ExampleFSM_Current() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	fmt.Println(fsm.Current())
	// Output: closed
}

func ExampleFSM_Is() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	fmt.Println(fsm.Is("closed"))
	fmt.Println(fsm.Is("open"))
	// Output:
	// true
	// false
}

func ExampleFSM_Can() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	fmt.Println(fsm.Can("open"))
	fmt.Println(fsm.Can("close"))
	// Output:
	// true
	// false
}

func ExampleFSM_AvailableTransitions() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
			{Name: "kick", Src: []string{"closed"}, Dst: "broken"},
		},
		Callbacks{},
	)
	// sort the results ordering is consistent for the output checker
	transitions := fsm.AvailableTransitions()
	sort.Strings(transitions)
	fmt.Println(transitions)
	// Output:
	// [kick open]
}

func ExampleFSM_Cannot() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	fmt.Println(fsm.Cannot("open"))
	fmt.Println(fsm.Cannot("close"))
	// Output:
	// false
	// true
}

func ExampleFSM_Event() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	fmt.Println(fsm.Current())
	err := fsm.Event(context.Background(), "open")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	err = fsm.Event(context.Background(), "close")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// closed
	// open
	// closed
}

func ExampleFSM_Transition() {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{
			"leave_closed": func(_ context.Context, e *Event) {
				e.Async()
			},
		},
	)
	err := fsm.Event(context.Background(), "open")
	if e, ok := err.(AsyncError); !ok && e.Err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	err = fsm.Transition()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// closed
	// open
}

func TestEventAndCanInGoroutines(t *testing.T) {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				_ = fsm.Event(context.Background(), "open")
			} else {
				_ = fsm.Event(context.Background(), "close")
			}
		}(i)
		go func() {
			defer wg.Done()
			fsm.Can("close")
		}()
	}
	wg.Wait()
}
