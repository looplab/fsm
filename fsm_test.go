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
	"fmt"
	"testing"
)

func TestSameState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "start"},
		},
		Callbacks{},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
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
	err := fsm.Event("close")
	if err.Error() != "event close inappropriate in current state closed" {
		t.FailNow()
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
	err := fsm.Event("lock")
	if err.Error() != "event lock does not exist" {
		t.FailNow()
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

	fsm.Event("first")
	if fsm.Current() != "two" {
		t.FailNow()
	}
	fsm.Event("reset")
	if fsm.Current() != "one" {
		t.FailNow()
	}
	fsm.Event("first")
	fsm.Event("second")
	if fsm.Current() != "three" {
		t.FailNow()
	}
	fsm.Event("reset")
	if fsm.Current() != "one" {
		t.FailNow()
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

	fsm.Event("first")
	fsm.Event("reset")
	if fsm.Current() != "reset_one" {
		t.FailNow()
	}
	fsm.Event("reset")
	if fsm.Current() != "start" {
		t.FailNow()
	}

	fsm.Event("second")
	fsm.Event("reset")
	if fsm.Current() != "reset_two" {
		t.FailNow()
	}
	fsm.Event("reset")
	if fsm.Current() != "start" {
		t.FailNow()
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
			"before_event": func(e *Event) {
				beforeEvent = true
			},
			"leave_state": func(e *Event) {
				leaveState = true
			},
			"enter_state": func(e *Event) {
				enterState = true
			},
			"after_event": func(e *Event) {
				afterEvent = true
			},
		},
	)

	fsm.Event("run")
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.FailNow()
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
			"before_run": func(e *Event) {
				beforeEvent = true
			},
			"leave_start": func(e *Event) {
				leaveState = true
			},
			"enter_end": func(e *Event) {
				enterState = true
			},
			"after_run": func(e *Event) {
				afterEvent = true
			},
		},
	)

	fsm.Event("run")
	if !(beforeEvent && leaveState && enterState && afterEvent) {
		t.FailNow()
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
			"end": func(e *Event) {
				enterState = true
			},
			"run": func(e *Event) {
				afterEvent = true
			},
		},
	)

	fsm.Event("run")
	if !(enterState && afterEvent) {
		t.FailNow()
	}
}

func TestCancelBeforeGenericEvent(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_event": func(e *Event) {
				e.Cancel()
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
	}
}

func TestCancelBeforeSpecificEvent(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_run": func(e *Event) {
				e.Cancel()
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
	}
}

func TestCancelLeaveGenericState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_state": func(e *Event) {
				e.Cancel()
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
	}
}

func TestCancelLeaveSpecificState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_start": func(e *Event) {
				e.Cancel()
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
	}
}

func TestAsyncTransitionGenericState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_state": func(e *Event) {
				e.Async()
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
	}
	fsm.Transition()
	if fsm.Current() != "end" {
		t.FailNow()
	}
}

func TestAsyncTransitionSpecificState(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"leave_start": func(e *Event) {
				e.Async()
			},
		},
	)
	fsm.Event("run")
	if fsm.Current() != "start" {
		t.FailNow()
	}
	fsm.Transition()
	if fsm.Current() != "end" {
		t.FailNow()
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
			"leave_start": func(e *Event) {
				e.Async()
			},
		},
	)
	fsm.Event("run")
	err := fsm.Event("reset")
	if err.Error() != "event reset inappropriate because previous transition did not complete" {
		t.FailNow()
	}
	fsm.Transition()
	fsm.Event("reset")
	if fsm.Current() != "start" {
		t.FailNow()
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
	if err.Error() != "transition inappropriate because no state change in progress" {
		t.FailNow()
	}
}

func TestCallbackNoError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(e *Event) {
			},
		},
	)
	e := fsm.Event("run")
	if e != nil {
		t.FailNow()
	}
}

func TestCallbackError(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(e *Event) {
				e.Err = fmt.Errorf("error")
			},
		},
	)
	e := fsm.Event("run")
	if e.Error() != "error" {
		t.FailNow()
	}
}

func TestCallbackArgs(t *testing.T) {
	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"run": func(e *Event) {
				if len(e.Args) != 1 {
					t.Fatal("too few arguments")
				}
				arg, ok := e.Args[0].(string)
				if !ok {
					t.Fatal("not a string argument")
				}
				if arg != "test" {
					t.Fatal("incorrect argument")
				}
			},
		},
	)
	fsm.Event("run", "test")
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
			"before_warn": func(e *Event) {
				fmt.Println("before_warn")
			},
			"before_event": func(e *Event) {
				fmt.Println("before_event")
			},
			"leave_green": func(e *Event) {
				fmt.Println("leave_green")
			},
			"leave_state": func(e *Event) {
				fmt.Println("leave_state")
			},
			"enter_yellow": func(e *Event) {
				fmt.Println("enter_yellow")
			},
			"enter_state": func(e *Event) {
				fmt.Println("enter_state")
			},
			"after_warn": func(e *Event) {
				fmt.Println("after_warn")
			},
			"after_event": func(e *Event) {
				fmt.Println("after_event")
			},
		},
	)
	fmt.Println(fsm.Current())
	err := fsm.Event("warn")
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
	err := fsm.Event("open")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	err = fsm.Event("close")
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
			"leave_closed": func(e *Event) {
				e.Async()
			},
		},
	)
	err := fsm.Event("open")
	if err != nil {
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
