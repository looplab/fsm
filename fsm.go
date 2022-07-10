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
//
package fsm

import (
	"fmt"
	"sync"

	"golang.org/x/exp/constraints"
)

// transitioner is an interface for the FSM's transition function.
type transitioner[E constraints.Ordered, S constraints.Ordered] interface {
	transition(*FSM[E, S]) error
}

// FSM is the state machine that holds the current state.
// E ist the event
// S is the state
// It has to be created with New to function properly.
type FSM[E constraints.Ordered, S constraints.Ordered] struct {
	// current is the state that the FSM is currently in.
	current S

	// transitions maps events and source states to destination states.
	transitions map[eKey[E, S]]S

	// callbacks maps events and targets to callback functions.
	callbacks Callbacks[E, S]

	// transition is the internal transition functions used either directly
	// or when Transition is called in an asynchronous state transition.
	transition func()
	// transitioner calls the FSM's transition() function.
	transitioner transitioner[E, S]

	// stateMu guards access to the current state.
	stateMu sync.RWMutex
	// eventMu guards access to Event() and Transition().
	eventMu sync.Mutex

	// metadata can be used to store and load data that maybe used across events
	// use methods SetMetadata() and Metadata() to store and load data
	metadata map[string]any
	// metadataMu guards access to the metadata.
	metadataMu sync.RWMutex
}

// Transition represents an event when initializing the FSM.
//
// The event can have one or more source states that is valid for performing
// the transition. If the FSM is in one of the source states it will end up in
// the specified destination state, calling all defined callbacks as it goes.
type Transition[E constraints.Ordered, S constraints.Ordered] struct {
	// Event is the event used when calling for a transition.
	Event E

	// Src is a slice of source states that the FSM must be in to perform a
	// state transition.
	Src []S

	// Dst is the destination state that the FSM will be in if the transition
	// succeeds.
	Dst S
}

// Transitions is a shorthand for defining the transition map in NewFSM.
type Transitions[E constraints.Ordered, S constraints.Ordered] []Transition[E, S]

// New constructs a generic FSM with a initial state S, for events E.
// E is the event type, S is the state type.
//
// Transistions define the state transistions that can be performed for a given event
// and a slice of source states, the destination state and the callback function.
//
// Callbacks are added as a slice specified as Callbacks and called in the same order.
func New[E constraints.Ordered, S constraints.Ordered](initial S, transitions Transitions[E, S], callbacks Callbacks[E, S]) *FSM[E, S] {
	f := &FSM[E, S]{
		current:      initial,
		transitioner: &defaultTransitioner[E, S]{},
		transitions:  map[eKey[E, S]]S{},
		callbacks:    callbacks,
		metadata:     map[string]any{},
	}

	// Build transition map and store sets of all events and states.
	for _, e := range transitions {
		for _, src := range e.Src {
			// FIXME eKey still required?
			f.transitions[eKey[E, S]{e.Event, src}] = e.Dst
		}
	}
	return f
}

// Current returns the current state of the FSM.
func (f *FSM[E, S]) Current() S {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return f.current
}

// Is returns true if state is the current state.
func (f *FSM[E, S]) Is(state S) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return state == f.current
}

// SetState allows the user to move to the given state from current state.
// The call does not trigger any callbacks, if defined.
func (f *FSM[E, S]) SetState(state S) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.current = state
}

// Can returns true if event can occur in the current state.
func (f *FSM[E, S]) Can(event E) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	_, ok := f.transitions[eKey[E, S]{event, f.current}]
	return ok && (f.transition == nil)
}

// Cannot returns true if event can not occur in the current state.
// It is a convenience method to help code read nicely.
func (f *FSM[E, S]) Cannot(event E) bool {
	return !f.Can(event)
}

// AvailableTransitions returns a list of transitions available in the
// current state.
func (f *FSM[E, S]) AvailableTransitions() []E {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	var transitions []E
	for key := range f.transitions {
		if key.src == f.current {
			transitions = append(transitions, key.event)
		}
	}
	return transitions
}

// Metadata returns the value stored in metadata
func (f *FSM[E, S]) Metadata(key string) (any, bool) {
	f.metadataMu.RLock()
	defer f.metadataMu.RUnlock()
	dataElement, ok := f.metadata[key]
	return dataElement, ok
}

// SetMetadata stores the dataValue in metadata indexing it with key
func (f *FSM[E, S]) SetMetadata(key string, value any) {
	f.metadataMu.Lock()
	defer f.metadataMu.Unlock()
	f.metadata[key] = value
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
func (f *FSM[E, S]) Event(event E, args ...any) error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()

	f.stateMu.RLock()
	defer f.stateMu.RUnlock()

	if f.transition != nil {
		return InTransitionError{fmt.Sprintf("%v", event)}
	}

	dst, ok := f.transitions[eKey[E, S]{event, f.current}]
	if !ok {
		for ekey := range f.transitions {
			if ekey.event == event {
				return InvalidEventError{fmt.Sprintf("%v", event), fmt.Sprintf("%v", f.current)}
			}
		}
		return UnknownEventError{fmt.Sprintf("%v", event)}
	}

	e := &CallbackContext[E, S]{f, event, f.current, dst, nil, args, false, false}

	err := f.beforeEventCallbacks(e)
	if err != nil {
		return err
	}

	if f.current == dst {
		f.afterEventCallbacks(e)
		return NoTransitionError{e.Err}
	}

	// Setup the transition, call it later.
	f.transition = func() {
		f.stateMu.Lock()
		f.current = dst
		f.stateMu.Unlock()

		f.enterStateCallbacks(e)
		f.afterEventCallbacks(e)
	}

	if err = f.leaveStateCallbacks(e); err != nil {
		if _, ok := err.(CanceledError); ok {
			f.transition = nil
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
func (f *FSM[E, S]) Transition() error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()
	return f.doTransition()
}

// doTransition wraps transitioner.transition.
func (f *FSM[E, S]) doTransition() error {
	return f.transitioner.transition(f)
}

// defaultTransitioner is the default implementation of the transitioner
// interface. Other implementations can be swapped in for testing.
type defaultTransitioner[E constraints.Ordered, S constraints.Ordered] struct{}

// Transition completes an asynchronous state change.
//
// The callback for leave_<STATE> must previously have called Async on its
// event to have initiated an asynchronous state transition.
func (t defaultTransitioner[E, S]) transition(f *FSM[E, S]) error {
	if f.transition == nil {
		return NotInTransitionError{}
	}
	f.transition()
	f.transition = nil
	return nil
}

// beforeEventCallbacks calls the before_ callbacks, first the named then the
// general version.
func (f *FSM[E, S]) beforeEventCallbacks(e *CallbackContext[E, S]) error {
	for _, cb := range f.callbacks {
		if cb.When == BeforeEvent {
			if cb.Event == e.Event {
				cb.F(e)
				if e.canceled {
					return CanceledError{e.Err}
				}
			}
		}
		if cb.When == BeforeAllEvents {
			cb.F(e)
			if e.canceled {
				return CanceledError{e.Err}
			}
		}
	}
	return nil
}

// leaveStateCallbacks calls the leave_ callbacks, first the named then the
// general version.
func (f *FSM[E, S]) leaveStateCallbacks(e *CallbackContext[E, S]) error {
	for _, cb := range f.callbacks {
		if cb.When == LeaveState {
			if cb.State == e.Src {
				cb.F(e)
				if e.canceled {
					return CanceledError{e.Err}
				} else if e.async {
					return AsyncError{e.Err}
				}
			}
		}
		if cb.When == LeaveAllStates {
			cb.F(e)
			if e.canceled {
				return CanceledError{e.Err}
			} else if e.async {
				return AsyncError{e.Err}
			}
		}
	}
	return nil
}

// enterStateCallbacks calls the enter_ callbacks, first the named then the
// general version.
func (f *FSM[E, S]) enterStateCallbacks(e *CallbackContext[E, S]) {
	for _, cb := range f.callbacks {
		if cb.When == EnterState {
			if cb.State == e.Dst {
				cb.F(e)
			}
		}
		if cb.When == EnterAllStates {
			cb.F(e)
		}
	}
}

// afterEventCallbacks calls the after_ callbacks, first the named then the
// general version.
func (f *FSM[E, S]) afterEventCallbacks(e *CallbackContext[E, S]) {
	for _, cb := range f.callbacks {
		if cb.When == AfterEvent {
			if cb.Event == e.Event {
				cb.F(e)
			}
		}
		if cb.When == AfterAllEvents {
			cb.F(e)
		}
	}
}

// eKey is a struct key used for storing the transition map.
type eKey[E constraints.Ordered, S constraints.Ordered] struct {
	// event is the name of the event that the keys refers to.
	event E

	// src is the source from where the event can transition.
	src S
}
