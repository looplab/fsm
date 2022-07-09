package main

import (
	"fmt"

	"github.com/looplab/fsm"
)

type MyEvent string
type MyState string

const (
	Close MyEvent = "close"
	Open  MyEvent = "open"
	Any   MyEvent = ""

	IsClosed MyState = "closed"
	IsOpen   MyState = "open"
)

func main() {
	fsm := fsm.New(
		IsClosed,
		fsm.Transitions[MyEvent, MyState]{
			{Event: Open, Src: []MyState{IsClosed}, Dst: IsOpen},
			{Event: Close, Src: []MyState{IsOpen}, Dst: IsClosed},
		},
		fsm.Callbacks[MyEvent, MyState]{
			fsm.Callback[MyEvent, MyState]{
				When: fsm.AfterEvent, Event: Open,
				F: func(cr *fsm.CallbackContext[MyEvent, MyState]) {
					fmt.Printf("callback: event:%s src:%s dst:%s\n", cr.Event, cr.Src, cr.Dst)
				},
			},
			fsm.Callback[MyEvent, MyState]{
				When: fsm.AfterAllEvents,
				F: func(cr *fsm.CallbackContext[MyEvent, MyState]) {
					fmt.Printf("callback after all: event:%s src:%s dst:%s\n", cr.Event, cr.Src, cr.Dst)
				},
			},
		},
	)
	fmt.Println(fsm.Current())
	err := fsm.Event(Open)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	err = fsm.Event(Close)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(fsm.Current())
	// Output:
	// closed
	// open
	// closed
}