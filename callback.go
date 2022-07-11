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
	"golang.org/x/exp/constraints"
)

// CallbackType defines at which type of Event this callback should be called.
type CallbackType int

const (
	// BeforeEvent called before event E
	BeforeEvent CallbackType = iota
	// BeforeAllEvents called before all events
	BeforeAllEvents
	// AfterEvent called after event E
	AfterEvent
	// AfterAllEvents called after all events
	AfterAllEvents
	// EnterState called after entering state S
	EnterState
	// EnterAllStates called after entering all states
	EnterAllStates
	// LeaveState is called before leaving state S.
	LeaveState
	// LeaveAllStates is called before leaving all states.
	LeaveAllStates
)

// Callback defines a condition when the callback function F should be called in certain conditions.
// The order of execution for CallbackTypes in the same event or state is:
// The concrete CallbackType has precedence over a general one, e.g.
// BeforEvent E will be fired before BeforeAllEvents.
type Callback[E constraints.Ordered, S constraints.Ordered] struct {
	// When should the callback be called.
	When CallbackType
	// Event is the event that the callback should be called for. Only relevant for BeforeEvent and AfterEvent.
	Event E
	// State is the state that the callback should be called for. Only relevant for EnterState and LeaveState.
	State S
	// F is the callback function.
	F func(*CallbackContext[E, S])
}

// Callbacks is a shorthand for defining the callbacks in New.
type Callbacks[E constraints.Ordered, S constraints.Ordered] []Callback[E, S]

// CallbackContext is the info that get passed as a reference in the callbacks.
type CallbackContext[E constraints.Ordered, S constraints.Ordered] struct {
	// FSM is an reference to the current FSM.
	FSM *FSM[E, S]
	// Event is the event name.
	Event E
	// Src is the state before the transition.
	Src S
	// Dst is the state after the transition.
	Dst S
	// Err is an optional error that can be returned from a callback.
	Err error
	// Args is an optional list of arguments passed to the callback.
	Args []any
	// canceled is an internal flag set if the transition is canceled.
	canceled bool
	// async is an internal flag set if the transition should be asynchronous
	async bool
}

// Cancel can be called in before_<EVENT> or leave_<STATE> to cancel the
// current transition before it happens. It takes an optional error, which will
// overwrite e.Err if set before.
func (ctx *CallbackContext[E, S]) Cancel(err ...error) {
	ctx.canceled = true

	if len(err) > 0 {
		ctx.Err = err[0]
	}
}

// Async can be called in leave_<STATE> to do an asynchronous state transition.
//
// The current state transition will be on hold in the old state until a final
// call to Transition is made. This will complete the transition and possibly
// call the other callbacks.
func (ctx *CallbackContext[E, S]) Async() {
	ctx.async = true
}
