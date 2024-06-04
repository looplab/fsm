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

// Package fsm implements a finite state machine.
//
// It is heavily based on two FSM implementations:
//
// Javascript Finite State Machine
// https://github.com/jakesgordon/javascript-state-machine
//
// Fysom for Python
// https://github.com/oxplot/fysom (forked at https://github.com/mriehl/fysom)
package fsm

import (
	"context"
	"strings"
	"sync"
)

// transitioner is an interface for the FSM's transition function.
type transitioner interface {
	transition(*FSM) error
}

// FSM is the state machine that holds the current state.
//
// It has to be created with NewFSM to function properly.
type FSM struct {
	// current is the state that the FSM is currently in.
	current string

	// transitions maps events and source states to destination states.
	transitions map[eKey]string

	// callbacks maps events and targets to callback functions.
	callbacks map[cKey]Callback

	// transition is the internal transition functions used either directly
	// or when Transition is called in an asynchronous state transition.
	transition func()
	// transitionerObj calls the FSM's transition() function.
	transitionerObj transitioner

	// stateMu guards access to the current state.
	stateMu sync.RWMutex
	// eventMu guards access to Event() and Transition().
	eventMu sync.Mutex
	// metadata can be used to store and load data that maybe used across events
	// use methods SetMetadata() and Metadata() to store and load data
	metadata map[string]interface{}

	metadataMu sync.RWMutex

	// processNoTransitionStates is used to allow events with no transitions 
	// to be performed
	processNoTransitionStates bool
}

// EventDesc represents an event when initializing the FSM.
//
// The event can have one or more source states that is valid for performing
// the transition. If the FSM is in one of the source states it will end up in
// the specified destination state, calling all defined callbacks as it goes.
type EventDesc struct {
	// Name is the event name used when calling for a transition.
	Name string

	// Src is a slice of source states that the FSM must be in to perform a
	// state transition.
	Src []string

	// Dst is the destination state that the FSM will be in if the transition
	// succeeds.
	Dst string
}

// Callback is a function type that callbacks should use. Event is the current
// event info as the callback happens.
type Callback func(context.Context, *Event)

// Events is a shorthand for defining the transition map in NewFSM.
type Events []EventDesc

// Callbacks is a shorthand for defining the callbacks in NewFSM.
type Callbacks map[string]Callback

// NewFSM constructs a FSM from events and callbacks.
//
// The events and transitions are specified as a slice of Event structs
// specified as Events. Each Event is mapped to one or more internal
// transitions from Event.Src to Event.Dst.
//
// Callbacks are added as a map specified as Callbacks where the key is parsed
// as the callback event as follows, and called in the same order:
//
// 1. before_<EVENT> - called before event named <EVENT>
//
// 2. before_event - called before all events
//
// 3. leave_<OLD_STATE> - called before leaving <OLD_STATE>
//
// 4. leave_state - called before leaving all states
//
// 5. enter_<NEW_STATE> - called after entering <NEW_STATE>
//
// 6. enter_state - called after entering all states
//
// 7. after_<EVENT> - called after event named <EVENT>
//
// 8. after_event - called after all events
//
// There are also two short form versions for the most commonly used callbacks.
// They are simply the name of the event or state:
//
// 1. <NEW_STATE> - called after entering <NEW_STATE>
//
// 2. <EVENT> - called after event named <EVENT>
//
// If both a shorthand version and a full version is specified it is undefined
// which version of the callback will end up in the internal map. This is due
// to the pseudo random nature of Go maps. No checking for multiple keys is
// currently performed.
func NewFSM(initial string, events []EventDesc, callbacks map[string]Callback) *FSM {
	f := &FSM{
		transitionerObj: &transitionerStruct{},
		current:         initial,
		transitions:     make(map[eKey]string),
		callbacks:       make(map[cKey]Callback),
		metadata:        make(map[string]interface{}),
		processNoTransitionStates: false,
	}

	// Build transition map and store sets of all events and states.
	allEvents := make(map[string]bool)
	allStates := make(map[string]bool)
	for _, e := range events {
		for _, src := range e.Src {
			f.transitions[eKey{e.Name, src}] = e.Dst
			allStates[src] = true
			allStates[e.Dst] = true
		}
		allEvents[e.Name] = true
	}

	// Map all callbacks to events/states.
	for name, fn := range callbacks {
		var target string
		var callbackType int

		switch {
		case strings.HasPrefix(name, "before_"):
			target = strings.TrimPrefix(name, "before_")
			if target == "event" {
				target = ""
				callbackType = callbackBeforeEvent
			} else if _, ok := allEvents[target]; ok {
				callbackType = callbackBeforeEvent
			}
		case strings.HasPrefix(name, "leave_"):
			target = strings.TrimPrefix(name, "leave_")
			if target == "state" {
				target = ""
				callbackType = callbackLeaveState
			} else if _, ok := allStates[target]; ok {
				callbackType = callbackLeaveState
			}
		case strings.HasPrefix(name, "enter_"):
			target = strings.TrimPrefix(name, "enter_")
			if target == "state" {
				target = ""
				callbackType = callbackEnterState
			} else if _, ok := allStates[target]; ok {
				callbackType = callbackEnterState
			}
		case strings.HasPrefix(name, "after_"):
			target = strings.TrimPrefix(name, "after_")
			if target == "event" {
				target = ""
				callbackType = callbackAfterEvent
			} else if _, ok := allEvents[target]; ok {
				callbackType = callbackAfterEvent
			}
		default:
			target = name
			if _, ok := allStates[target]; ok {
				callbackType = callbackEnterState
			} else if _, ok := allEvents[target]; ok {
				callbackType = callbackAfterEvent
			}
		}

		if callbackType != callbackNone {
			f.callbacks[cKey{target, callbackType}] = fn
		}
	}

	return f
}

// Current returns the current state of the FSM.
func (f *FSM) Current() string {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return f.current
}

// Is returns true if state is the current state.
func (f *FSM) Is(state string) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return state == f.current
}

// SetState allows the user to move to the given state from current state.
// The call does not trigger any callbacks, if defined.
func (f *FSM) SetState(state string) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.current = state
}

// Can returns true if event can occur in the current state.
func (f *FSM) Can(event string) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	_, ok := f.transitions[eKey{event, f.current}]
	return ok && (f.transition == nil)
}

// AvailableTransitions returns a list of transitions available in the
// current state.
func (f *FSM) AvailableTransitions() []string {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	var transitions []string
	for key := range f.transitions {
		if key.src == f.current {
			transitions = append(transitions, key.event)
		}
	}
	return transitions
}

// Cannot returns true if event can not occur in the current state.
// It is a convenience method to help code read nicely.
func (f *FSM) Cannot(event string) bool {
	return !f.Can(event)
}

// Metadata returns the value stored in metadata
func (f *FSM) Metadata(key string) (interface{}, bool) {
	f.metadataMu.RLock()
	defer f.metadataMu.RUnlock()
	dataElement, ok := f.metadata[key]
	return dataElement, ok
}

// SetMetadata stores the dataValue in metadata indexing it with key
func (f *FSM) SetMetadata(key string, dataValue interface{}) {
	f.metadataMu.Lock()
	defer f.metadataMu.Unlock()
	f.metadata[key] = dataValue
}

// DeleteMetadata deletes the dataValue in metadata by key
func (f *FSM) DeleteMetadata(key string) {
	f.metadataMu.Lock()
	delete(f.metadata, key)
	f.metadataMu.Unlock()
}

// SetProcessNoTransitionStates allow the users to enable events with no transitions
// to be performed
func (f *FSM) SetProcessNoTransitionStates(process bool) {
	f.processNoTransitionStates = process
}

// Event initiates a state transition with the named event.
//
// The call takes a variable number of arguments that will be passed to the
// callback, if defined.
//
// It will return nil if the state change is ok or one of these errors:
//
// - event X inappropriate because previous transition did not complete
//
// - event X inappropriate in current state Y
//
// - event X does not exist
//
// - internal error on state transition
//
// The last error should never occur in this situation and is a sign of an
// internal bug.
func (f *FSM) Event(ctx context.Context, event string, args ...interface{}) error {
	f.eventMu.Lock()
	// in order to always unlock the event mutex, the defer is added
	// in case the state transition goes through and enter/after callbacks
	// are called; because these must be able to trigger new state
	// transitions, it is explicitly unlocked in the code below
	var unlocked bool
	defer func() {
		if !unlocked {
			f.eventMu.Unlock()
		}
	}()

	f.stateMu.RLock()
	defer f.stateMu.RUnlock()

	if f.transition != nil {
		return InTransitionError{event}
	}

	dst, ok := f.transitions[eKey{event, f.current}]
	if !ok {
		for ekey := range f.transitions {
			if ekey.event == event {
				return InvalidEventError{event, f.current}
			}
		}
		return UnknownEventError{event}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	e := &Event{f, event, f.current, dst, nil, args, false, false, cancel}

	err := f.beforeEventCallbacks(ctx, e)
	if err != nil {
		return err
	}

	if f.current == dst && !f.processNoTransitionStates {
		f.stateMu.RUnlock()
		defer f.stateMu.RLock()
		f.eventMu.Unlock()
		unlocked = true
		f.afterEventCallbacks(ctx, e)
		return NoTransitionError{e.Err}
	}

	// Setup the transition, call it later.
	transitionFunc := func(ctx context.Context, async bool) func() {
		return func() {
			if ctx.Err() != nil {
				if e.Err == nil {
					e.Err = ctx.Err()
				}
				return
			}

			f.stateMu.Lock()
			f.current = dst
			f.transition = nil // treat the state transition as done
			f.stateMu.Unlock()

			// at this point, we unlock the event mutex in order to allow
			// enter state callbacks to trigger another transition
			// for aynchronous state transitions this doesn't happen because
			// the event mutex has already been unlocked
			if !async {
				f.eventMu.Unlock()
				unlocked = true
			}
			f.enterStateCallbacks(ctx, e)
			f.afterEventCallbacks(ctx, e)
		}
	}

	f.transition = transitionFunc(ctx, false)

	if err = f.leaveStateCallbacks(ctx, e); err != nil {
		if _, ok := err.(CanceledError); ok {
			f.transition = nil
		} else if asyncError, ok := err.(AsyncError); ok {
			// setup a new context in order for async state transitions to work correctly
			// this "uncancels" the original context which ignores its cancelation
			// but keeps the values of the original context available to callers
			ctx, cancel := uncancelContext(ctx)
			e.cancelFunc = cancel
			asyncError.Ctx = ctx
			asyncError.CancelTransition = cancel
			f.transition = transitionFunc(ctx, true)
			return asyncError
		}
		return err
	}

	// Perform the rest of the transition, if not asynchronous.
	f.stateMu.RUnlock()
	defer f.stateMu.RLock()
	err = f.doTransition()
	if err != nil {
		return InternalError{}
	}

	return e.Err
}

// Transition wraps transitioner.transition.
func (f *FSM) Transition() error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()
	return f.doTransition()
}

// doTransition wraps transitioner.transition.
func (f *FSM) doTransition() error {
	return f.transitionerObj.transition(f)
}

// transitionerStruct is the default implementation of the transitioner
// interface. Other implementations can be swapped in for testing.
type transitionerStruct struct{}

// Transition completes an asynchronous state change.
//
// The callback for leave_<STATE> must previously have called Async on its
// event to have initiated an asynchronous state transition.
func (t transitionerStruct) transition(f *FSM) error {
	if f.transition == nil {
		return NotInTransitionError{}
	}
	f.transition()
	return nil
}

// beforeEventCallbacks calls the before_ callbacks, first the named then the
// general version.
func (f *FSM) beforeEventCallbacks(ctx context.Context, e *Event) error {
	if fn, ok := f.callbacks[cKey{e.Event, callbackBeforeEvent}]; ok {
		fn(ctx, e)
		if e.canceled {
			return CanceledError{e.Err}
		}
	}
	if fn, ok := f.callbacks[cKey{"", callbackBeforeEvent}]; ok {
		fn(ctx, e)
		if e.canceled {
			return CanceledError{e.Err}
		}
	}
	return nil
}

// leaveStateCallbacks calls the leave_ callbacks, first the named then the
// general version.
func (f *FSM) leaveStateCallbacks(ctx context.Context, e *Event) error {
	if fn, ok := f.callbacks[cKey{f.current, callbackLeaveState}]; ok {
		fn(ctx, e)
		if e.canceled {
			return CanceledError{e.Err}
		} else if e.async {
			return AsyncError{Err: e.Err}
		}
	}
	if fn, ok := f.callbacks[cKey{"", callbackLeaveState}]; ok {
		fn(ctx, e)
		if e.canceled {
			return CanceledError{e.Err}
		} else if e.async {
			return AsyncError{Err: e.Err}
		}
	}
	return nil
}

// enterStateCallbacks calls the enter_ callbacks, first the named then the
// general version.
func (f *FSM) enterStateCallbacks(ctx context.Context, e *Event) {
	if fn, ok := f.callbacks[cKey{f.current, callbackEnterState}]; ok {
		fn(ctx, e)
	}
	if fn, ok := f.callbacks[cKey{"", callbackEnterState}]; ok {
		fn(ctx, e)
	}
}

// afterEventCallbacks calls the after_ callbacks, first the named then the
// general version.
func (f *FSM) afterEventCallbacks(ctx context.Context, e *Event) {
	if fn, ok := f.callbacks[cKey{e.Event, callbackAfterEvent}]; ok {
		fn(ctx, e)
	}
	if fn, ok := f.callbacks[cKey{"", callbackAfterEvent}]; ok {
		fn(ctx, e)
	}
}

const (
	callbackNone int = iota
	callbackBeforeEvent
	callbackLeaveState
	callbackEnterState
	callbackAfterEvent
)

// cKey is a struct key used for keeping the callbacks mapped to a target.
type cKey struct {
	// target is either the name of a state or an event depending on which
	// callback type the key refers to. It can also be "" for a non-targeted
	// callback like before_event.
	target string

	// callbackType is the situation when the callback will be run.
	callbackType int
}

// eKey is a struct key used for storing the transition map.
type eKey struct {
	// event is the name of the event that the keys refers to.
	event string

	// src is the source from where the event can transition.
	src string
}
