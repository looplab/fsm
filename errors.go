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

type InvalidEventError struct {
	Event string
	State string
}

func (e *InvalidEventError) Error() string {
	return "event " + e.Event + " inappropriate in current state " + e.State
}

type UnknownEventError struct {
	Event string
}

func (e *UnknownEventError) Error() string {
	return "event " + e.Event + " does not exist"
}

type InTransitionError struct {
	Event string
}

func (e *InTransitionError) Error() string {
	return "event " + e.Event + " inappropriate because previous transition did not complete"
}

type NotInTransitionError struct {
}

func (e *NotInTransitionError) Error() string {
	return "transition inappropriate because no state change in progress"
}

type NoTransitionError struct {
	Err error
}

func (e *NoTransitionError) Error() string {
	if e.Err != nil {
		return "no transition with error: " + e.Err.Error()
	}
	return "no transition"
}

type CanceledError struct {
	Err error
}

func (e *CanceledError) Error() string {
	if e.Err != nil {
		return "transition canceled with error: " + e.Err.Error()
	}
	return "transition canceled"
}

type AsyncError struct {
	Err error
}

func (e *AsyncError) Error() string {
	if e.Err != nil {
		return "async started with error: " + e.Err.Error()
	}
	return "async started"
}

type InternalError struct {
}

func (e *InternalError) Error() string {
	return "internal error on state transition"
}
