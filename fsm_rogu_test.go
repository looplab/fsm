package fsm

import (
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
