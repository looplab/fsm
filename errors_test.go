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
	"errors"
	"testing"
)

func TestInvalidEventError(t *testing.T) {
	event := "invalid event"
	state := "state"
	e := InvalidEventError{Event: event, State: state}
	if e.Error() != "event "+e.Event+" inappropriate in current state "+e.State {
		t.Error("InvalidEventError string mismatch")
	}
}

func TestUnknownEventError(t *testing.T) {
	event := "invalid event"
	e := UnknownEventError{Event: event}
	if e.Error() != "event "+e.Event+" does not exist" {
		t.Error("UnknownEventError string mismatch")
	}
}

func TestInTransitionError(t *testing.T) {
	event := "in transition"
	e := InTransitionError{Event: event}
	if e.Error() != "event "+e.Event+" inappropriate because previous transition did not complete" {
		t.Error("InTransitionError string mismatch")
	}
}

func TestNotInTransitionError(t *testing.T) {
	e := NotInTransitionError{}
	if e.Error() != "transition inappropriate because no state change in progress" {
		t.Error("NotInTransitionError string mismatch")
	}
}

func TestNoTransitionError(t *testing.T) {
	e := NoTransitionError{}
	innerErr := errors.New("no transition")
	if e.Error() != "no transition" {
		t.Error("NoTransitionError string mismatch")
	}
	e.Err = innerErr
	if e.Error() != "no transition with error: "+e.Err.Error() {
		t.Error("NoTransitionError string mismatch")
	}

	realErr := hideErrInterfaceType(e)
	if !errors.Is(realErr, NoTransitionError{}) {
		t.Error("NoTransitionError 'Is' broken")
	}
	if !errors.Is(realErr, innerErr) {
		t.Error("NoTransitionError 'Is' broken")
	}
	if errors.Unwrap(e) != innerErr {
		t.Error("NoTransitionError 'Unwrap' broken")
	}
}

func TestCanceledError(t *testing.T) {
	e := CanceledError{}
	innerErr := errors.New("canceled")
	if e.Error() != "transition canceled" {
		t.Error("CanceledError string mismatch")
	}
	e.Err = innerErr
	if e.Error() != "transition canceled with error: "+e.Err.Error() {
		t.Error("CanceledError string mismatch")
	}

	realErr := hideErrInterfaceType(e)
	if !errors.Is(realErr, CanceledError{}) {
		t.Error("CanceledError 'Is' broken")
	}
	if !errors.Is(realErr, innerErr) {
		t.Error("CanceledError 'Is' broken")
	}
	if errors.Unwrap(e) != innerErr {
		t.Error("CanceledError 'Unwrap' broken")
	}
}

func TestAsyncError(t *testing.T) {
	e := AsyncError{}
	innerErr := errors.New("async")
	if e.Error() != "async started" {
		t.Error("AsyncError string mismatch")
	}
	e.Err = innerErr
	if e.Error() != "async started with error: "+e.Err.Error() {
		t.Error("AsyncError string mismatch")
	}

	realErr := hideErrInterfaceType(e)
	if !errors.Is(realErr, AsyncError{}) {
		t.Error("AsyncError 'Is' broken")
	}
	if !errors.Is(realErr, innerErr) {
		t.Error("AsyncError 'Is' broken")
	}
	if errors.Unwrap(e) != innerErr {
		t.Error("AsyncError 'Unwrap' broken")
	}
}

func TestInternalError(t *testing.T) {
	e := InternalError{}
	if e.Error() != "internal error on state transition" {
		t.Error("InternalError string mismatch")
	}
}

func hideErrInterfaceType(err error) error {
	return err
}
