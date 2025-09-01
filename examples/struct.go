//go:build ignore
// +build ignore

package main

import (
	"fmt"

	"github.com/looplab/fsm/v2"
)

type Door struct {
	To  string
	FSM *fsm.FSM[string, string]
}

func NewDoor(to string) *Door {
	d := &Door{
		To: to,
	}

	var err error
	d.FSM, err = fsm.New(
		"closed",
		fsm.Transitions[string, string]{
			{Event: "open", Src: []string{"closed"}, Dst: "open"},
			{Event: "close", Src: []string{"open"}, Dst: "closed"},
		},
		fsm.Callbacks[string, string]{
			fsm.Callback[string, string]{When: fsm.EnterAllStates,
				F: func(e *fsm.CallbackContext[string, string]) { d.enterState(e) },
			},
		},
	)
	if err != nil {
		fmt.Println(err)
	}
	return d
}

func (d *Door) enterState(e *fsm.CallbackContext[string, string]) {
	fmt.Printf("The door to %s is %s\n", d.To, e.Dst)
}

func main() {
	door := NewDoor("heaven")

	err := door.FSM.Event("open")
	if err != nil {
		fmt.Println(err)
	}

	err = door.FSM.Event("close")
	if err != nil {
		fmt.Println(err)
	}
}
