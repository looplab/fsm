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
	"reflect"
	"sort"
	"sync"
	"testing"
)

// runDifferential builds the same machine twice, once with NewFSM and once
// with NewFSMFromSpec from a Spec compiled by NewSpec, and drives both of them
// with drive. Callbacks and drive record everything that is observable into
// the log, so any difference in callback order, errors or states shows up as a
// difference between the two logs.
//
// Note that NewFSM is implemented with NewSpec and NewFSMFromSpec, so the two
// runs share their code path. These tests are a guard against the two
// constructors diverging later and coverage of the whole feature matrix
// through NewFSMFromSpec, they are not by themselves proof that the extraction
// of the Spec preserved behavior. That proof is the unmodified test suite that
// was here before it.
func runDifferential(t *testing.T, initial string, events Events, callbacks func(log *[]string) Callbacks, drive func(f *FSM, log *[]string)) {
	t.Helper()

	var fsmLog []string
	drive(NewFSM(initial, events, callbacks(&fsmLog)), &fsmLog)

	var specLog []string
	drive(NewFSMFromSpec(initial, NewSpec(events, callbacks(&specLog))), &specLog)

	if !reflect.DeepEqual(fsmLog, specLog) {
		t.Errorf("expected identical behavior, NewFSM gave %v, NewFSMFromSpec gave %v", fsmLog, specLog)
	}
}

// recordCallback returns a callback that records that it was called.
func recordCallback(log *[]string, name string) Callback {
	return func(_ context.Context, e *Event) {
		*log = append(*log, fmt.Sprintf("%s(%s:%s->%s)", name, e.Event, e.Src, e.Dst))
	}
}

// recordEvent records the outcome of an event, both the error and the state
// that the machine ended up in.
func recordEvent(f *FSM, log *[]string, event string, args ...interface{}) {
	err := f.Event(context.Background(), event, args...)
	*log = append(*log, fmt.Sprintf("event %s: err %T %v, state %s, transitions %v", event, err, err, f.Current(), sortedTransitions(f)))
}

func sortedTransitions(f *FSM) []string {
	transitions := f.AvailableTransitions()
	sort.Strings(transitions)
	return transitions
}

// richEvents is a machine with multiple events, multiple sources for one
// event and a transition back into the same state.
var richEvents = Events{
	{Name: "warn", Src: []string{"green"}, Dst: "yellow"},
	{Name: "panic", Src: []string{"green", "yellow"}, Dst: "red"},
	{Name: "calm", Src: []string{"red"}, Dst: "yellow"},
	{Name: "clear", Src: []string{"yellow"}, Dst: "green"},
	{Name: "stay", Src: []string{"green"}, Dst: "green"},
	{Name: "fail", Src: []string{"red"}, Dst: "broken"},
}

var richStates = []string{"green", "yellow", "red", "broken"}

// richCallbacks has a callback of every phase, in both the general and the
// named form, the two shorthand forms and one callback setting an error.
func richCallbacks(log *[]string) Callbacks {
	return Callbacks{
		"before_warn":  recordCallback(log, "before_warn"),
		"before_event": recordCallback(log, "before_event"),
		"leave_green":  recordCallback(log, "leave_green"),
		"leave_state":  recordCallback(log, "leave_state"),
		"enter_yellow": recordCallback(log, "enter_yellow"),
		"enter_state":  recordCallback(log, "enter_state"),
		"after_warn":   recordCallback(log, "after_warn"),
		"after_event":  recordCallback(log, "after_event"),
		"red":          recordCallback(log, "red"),   // shorthand for enter_red
		"clear":        recordCallback(log, "clear"), // shorthand for after_clear
		"after_fail": func(ctx context.Context, e *Event) {
			recordCallback(log, "after_fail")(ctx, e)
			e.Err = errors.New("failed")
		},
	}
}

func TestSpecSameBehaviorForAllTransitions(t *testing.T) {
	// All events of the machine plus one that it does not know about.
	events := []string{"warn", "panic", "calm", "clear", "stay", "fail", "unknown"}
	for _, state := range richStates {
		for _, event := range events {
			state, event := state, event
			t.Run(state+"_"+event, func(t *testing.T) {
				runDifferential(t, state, richEvents, richCallbacks, func(f *FSM, log *[]string) {
					recordEvent(f, log, event)
				})
			})
		}
	}
}

func TestSpecSameBehaviorForCallbackArgs(t *testing.T) {
	runDifferential(t, "green", richEvents, richCallbacks, func(f *FSM, log *[]string) {
		recordEvent(f, log, "warn", "with", "args")
	})
}

func TestSpecSameBehaviorForCancel(t *testing.T) {
	callbacks := func(log *[]string) Callbacks {
		return Callbacks{
			"before_event": func(ctx context.Context, e *Event) {
				recordCallback(log, "before_event")(ctx, e)
				e.Cancel()
			},
			"leave_start": func(ctx context.Context, e *Event) {
				recordCallback(log, "leave_start")(ctx, e)
				e.Cancel(errors.New("canceled by callback"))
			},
		}
	}
	events := Events{
		{Name: "run", Src: []string{"start"}, Dst: "end"},
		{Name: "walk", Src: []string{"start"}, Dst: "end"},
	}
	runDifferential(t, "start", events, callbacks, func(f *FSM, log *[]string) {
		recordEvent(f, log, "run")
	})

	// The same machine without the before_event callback, so that the cancel
	// with an error in leave_start is reached.
	leaveCallbacks := func(log *[]string) Callbacks {
		c := callbacks(log)
		delete(c, "before_event")
		return c
	}
	runDifferential(t, "start", events, leaveCallbacks, func(f *FSM, log *[]string) {
		recordEvent(f, log, "walk")
	})
}

func TestSpecSameBehaviorForAsyncTransition(t *testing.T) {
	events := Events{
		{Name: "run", Src: []string{"start"}, Dst: "end"},
		{Name: "reset", Src: []string{"end"}, Dst: "start"},
	}
	callbacks := func(log *[]string) Callbacks {
		return Callbacks{
			"leave_start": func(ctx context.Context, e *Event) {
				recordCallback(log, "leave_start")(ctx, e)
				e.Async()
			},
			"enter_end":   recordCallback(log, "enter_end"),
			"after_event": recordCallback(log, "after_event"),
		}
	}

	// Async transition completed with Transition().
	runDifferential(t, "start", events, callbacks, func(f *FSM, log *[]string) {
		recordEvent(f, log, "run")
		err := f.Transition()
		*log = append(*log, fmt.Sprintf("transition: err %T %v, state %s", err, err, f.Current()))
	})

	// An event issued while the async transition is in progress must give an
	// InTransitionError, and a Transition() without a started transition a
	// NotInTransitionError.
	runDifferential(t, "start", events, callbacks, func(f *FSM, log *[]string) {
		err := f.Transition()
		*log = append(*log, fmt.Sprintf("transition: err %T %v, state %s", err, err, f.Current()))
		recordEvent(f, log, "run")
		recordEvent(f, log, "reset")
		err = f.Transition()
		*log = append(*log, fmt.Sprintf("transition: err %T %v, state %s", err, err, f.Current()))
		recordEvent(f, log, "reset")
	})

	// The async transition canceled with AsyncError.CancelTransition.
	runDifferential(t, "start", events, callbacks, func(f *FSM, log *[]string) {
		err := f.Event(context.Background(), "run")
		asyncError, ok := err.(AsyncError)
		if !ok {
			t.Fatalf("expected error to be 'AsyncError', got %v", err)
		}
		canceled := make(chan struct{})
		go func() {
			<-asyncError.Ctx.Done()
			close(canceled)
		}()
		asyncError.CancelTransition()
		<-canceled

		err = f.Transition()
		*log = append(*log, fmt.Sprintf("transition: err %T %v, state %s", err, err, f.Current()))
	})
}

func TestSpecSameBehaviorForCanceledContext(t *testing.T) {
	events := Events{
		{Name: "run", Src: []string{"start"}, Dst: "end"},
		{Name: "finish", Src: []string{"end"}, Dst: "finished"},
	}
	callbacks := func(log *[]string) Callbacks {
		return Callbacks{
			"enter_end": func(ctx context.Context, e *Event) {
				recordCallback(log, "enter_end")(ctx, e)
				// cancel the context that was given to Event, from inside the
				// transition, so that the outcome is deterministic
				cancel, _ := e.FSM.Metadata("cancel")
				cancel.(context.CancelFunc)()
				<-ctx.Done()
				if err := e.FSM.Event(ctx, "finish"); err != nil {
					e.Err = fmt.Errorf("transitioning to the finished state failed: %w", err)
				}
			},
		}
	}
	runDifferential(t, "start", events, callbacks, func(f *FSM, log *[]string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		f.SetMetadata("cancel", context.CancelFunc(cancel))
		err := f.Event(ctx, "run")
		*log = append(*log, fmt.Sprintf("event run: canceled %t, err %v, state %s", errors.Is(err, context.Canceled), err, f.Current()))
	})
}

func TestSpecSameVisualization(t *testing.T) {
	fsm := NewFSM("yellow", richEvents, richCallbacks(new([]string)))
	shared := NewFSMFromSpec("yellow", NewSpec(richEvents, richCallbacks(new([]string))))

	for _, visualizeType := range []VisualizeType{GRAPHVIZ, MERMAID, MermaidStateDiagram, MermaidFlowChart} {
		expected, err := VisualizeWithType(fsm, visualizeType)
		if err != nil {
			t.Fatalf("visualization failed %v", err)
		}
		got, err := VisualizeWithType(shared, visualizeType)
		if err != nil {
			t.Fatalf("visualization failed %v", err)
		}
		if expected != got {
			t.Errorf("expected identical %s visualization, got %q and %q", visualizeType, expected, got)
		}
	}

	expected, err := VisualizeForMermaidWithGraphType(fsm, FlowChart)
	if err != nil {
		t.Fatalf("visualization failed %v", err)
	}
	got, err := VisualizeForMermaidWithGraphType(shared, FlowChart)
	if err != nil {
		t.Fatalf("visualization failed %v", err)
	}
	if expected != got {
		t.Errorf("expected identical %s visualization, got %q and %q", FlowChart, expected, got)
	}
}

func TestSpecIndependentState(t *testing.T) {
	spec := NewSpec(richEvents, nil)
	first := NewFSMFromSpec("green", spec)
	second := NewFSMFromSpec("green", spec)

	if err := first.Event(context.Background(), "warn"); err != nil {
		t.Errorf("transition failed %v", err)
	}
	if first.Current() != "yellow" {
		t.Error("expected state to be 'yellow'")
	}
	if second.Current() != "green" {
		t.Error("expected state to be 'green'")
	}

	second.SetState("red")
	if first.Current() != "yellow" {
		t.Error("expected state to be 'yellow'")
	}
	if second.Current() != "red" {
		t.Error("expected state to be 'red'")
	}
}

// assertNoTransitions checks that a machine without a Spec answers as a
// machine without any transitions instead of panicking.
func assertNoTransitions(t *testing.T, f *FSM) {
	t.Helper()

	if f.Can("run") {
		t.Error("expected event 'run' to not be possible")
	}
	if err := f.Event(context.Background(), "run"); err == nil {
		t.Error("expected 'UnknownEventError'")
	} else if _, ok := err.(UnknownEventError); !ok {
		t.Errorf("expected 'UnknownEventError', got %T %v", err, err)
	}
	if transitions := f.AvailableTransitions(); len(transitions) != 0 {
		t.Errorf("expected no available transitions, got %v", transitions)
	}
	if graph, expected := Visualize(f), "digraph fsm {\n\n}\n"; graph != expected {
		t.Errorf("expected graph to be %q, got %q", expected, graph)
	}
}

func TestZeroValueFSM(t *testing.T) {
	var f FSM
	assertNoTransitions(t, &f)
}

func TestNilSpec(t *testing.T) {
	assertNoTransitions(t, NewFSMFromSpec("green", nil))
}

func TestSpecIndependentMetadata(t *testing.T) {
	spec := NewSpec(richEvents, nil)
	first := NewFSMFromSpec("green", spec)
	second := NewFSMFromSpec("green", spec)

	// the metadata is only allocated when it is first written to
	if value, ok := first.Metadata("key"); ok {
		t.Errorf("expected no metadata, got %v", value)
	}
	first.DeleteMetadata("key")

	first.SetMetadata("key", "first")
	second.SetMetadata("key", "second")

	if value, ok := first.Metadata("key"); !ok || value != "first" {
		t.Errorf("expected metadata to be 'first', got %v", value)
	}
	if value, ok := second.Metadata("key"); !ok || value != "second" {
		t.Errorf("expected metadata to be 'second', got %v", value)
	}

	first.DeleteMetadata("key")
	if _, ok := first.Metadata("key"); ok {
		t.Error("expected metadata to be deleted")
	}
	if value, ok := second.Metadata("key"); !ok || value != "second" {
		t.Errorf("expected metadata to be 'second', got %v", value)
	}
}

func TestSpecSharedByConcurrentFSMs(t *testing.T) {
	spec := NewSpec(richEvents, Callbacks{
		"before_event": func(_ context.Context, e *Event) {},
		"enter_state":  func(_ context.Context, e *Event) {},
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fsm := NewFSMFromSpec("green", spec)
			for j := 0; j < 10; j++ {
				if err := fsm.Event(context.Background(), "warn"); err != nil {
					t.Error(err)
				}
				if err := fsm.Event(context.Background(), "clear"); err != nil {
					t.Error(err)
				}
			}
			if fsm.Current() != "green" {
				t.Error("expected state to be 'green'")
			}
		}()
	}
	wg.Wait()
}

func ExampleNewSpec() {
	spec := NewSpec(
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{
			"enter_state": func(_ context.Context, e *Event) {
				fmt.Println("door is", e.Dst)
			},
		},
	)

	first := NewFSMFromSpec("closed", spec)
	second := NewFSMFromSpec("open", spec)

	if err := first.Event(context.Background(), "open"); err != nil {
		fmt.Println(err)
	}
	if err := second.Event(context.Background(), "close"); err != nil {
		fmt.Println(err)
	}
	// Output:
	// door is open
	// door is closed
}

// benchFSM keeps the constructed machines alive so that the benchmarks below
// are not optimized away.
var benchFSM *FSM

func BenchmarkNewFSM(b *testing.B) {
	callbacks := richCallbacks(new([]string))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFSM = NewFSM("green", richEvents, callbacks)
	}
}

func BenchmarkNewFSMFromSpec(b *testing.B) {
	spec := NewSpec(richEvents, richCallbacks(new([]string)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchFSM = NewFSMFromSpec("green", spec)
	}
}
