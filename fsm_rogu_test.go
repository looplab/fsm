package fsm

import (
	"fmt"
	"log"
	"testing"
)

func TestGetCurrent(t *testing.T) {
	f := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "start", Msg: "Hi kaka!"},
		},
		Callbacks{},
	)

	msg := f.GetMessage("run", "start")
	if msg == "" {
		t.Error("message doesn't exist")
	}
}

func TestCases(t *testing.T) {
	f := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open", Msg: "Current State is Opened!"},
			{Name: "close", Src: []string{"open"}, Dst: "closed", Msg: "Current State is Closed!"},
		},
		Callbacks{},
	)

	fmt.Println(f.Current())

	err := f.Event("open")
	if err != nil {
		t.Error(err)
	}

	log.Println(f.Current())
	fmt.Println(f.GetMessage("open", f.Current()))

	err = f.Event("close")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(f.Current())
}
