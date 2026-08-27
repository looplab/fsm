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

import "strings"

// Spec is a compiled machine description, holding the transition map and the
// callbacks resolved from their names.
//
// It has to be created with NewSpec to function properly.
//
// A Spec is immutable once created and holds no state of its own, so a single
// Spec can be shared by any number of FSMs created with NewFSMFromSpec, also
// concurrently. This avoids recompiling the same tables for every machine when
// many machines share one description. A Spec must never be mutated after it
// has been created.
type Spec struct {
	// transitions maps events and source states to destination states.
	transitions map[eKey]string

	// callbacks maps events and targets to callback functions.
	callbacks map[cKey]Callback
}

// transitionFor returns the destination state for the event in the src state,
// and whether the transition exists. A nil Spec has no transitions.
func (s *Spec) transitionFor(event, src string) (string, bool) {
	if s == nil {
		return "", false
	}
	dst, ok := s.transitions[eKey{event, src}]
	return dst, ok
}

// transitionTable returns the transitions of the Spec, for reading only. A nil
// Spec gives a nil map, which is empty when read.
func (s *Spec) transitionTable() map[eKey]string {
	if s == nil {
		return nil
	}
	return s.transitions
}

// callbackFor returns the callback registered for the target of the given
// callback type, and whether it exists. A nil Spec has no callbacks.
func (s *Spec) callbackFor(target string, callbackType int) (Callback, bool) {
	if s == nil {
		return nil, false
	}
	fn, ok := s.callbacks[cKey{target, callbackType}]
	return fn, ok
}

// NewSpec compiles events and callbacks into a Spec that can be shared by any
// number of FSMs.
//
// The events and callbacks are interpreted exactly as described for NewFSM,
// which is implemented in terms of NewSpec.
func NewSpec(events []EventDesc, callbacks map[string]Callback) *Spec {
	s := &Spec{
		transitions: make(map[eKey]string),
		callbacks:   make(map[cKey]Callback),
	}

	// Build transition map and store sets of all events and states.
	allEvents := make(map[string]bool)
	allStates := make(map[string]bool)
	for _, e := range events {
		for _, src := range e.Src {
			s.transitions[eKey{e.Name, src}] = e.Dst
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
			s.callbacks[cKey{target, callbackType}] = fn
		}
	}

	return s
}

// NewFSMFromSpec constructs a FSM in the initial state from an already
// compiled Spec.
//
// The FSM shares the Spec with any other FSM created from it, everything that
// changes while the machine runs is kept per FSM.
//
// The initial state should be one of the states of the Spec. An unknown state
// is tolerated, the machine then has no transitions available until SetState
// is called and the Mermaid flow chart visualization highlights a node that is
// not part of the graph, the same as with NewFSM.
//
// A nil Spec is also tolerated, it gives a machine without transitions and
// callbacks that answers every event with an UnknownEventError.
func NewFSMFromSpec(initial string, spec *Spec) *FSM {
	return &FSM{
		transitionerObj: &transitionerStruct{},
		current:         initial,
		spec:            spec,
	}
}
