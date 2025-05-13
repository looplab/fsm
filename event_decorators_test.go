package fsm

import (
	"context"
	"errors"
	"testing"
)

func TestDecorateCallbackWithErrorHandling(t *testing.T) {
	t.Parallel()

	fsm := NewFSM(
		"start",
		Events{
			{Name: "run", Src: []string{"start"}, Dst: "end"},
		},
		Callbacks{
			"before_event": DecorateCallbackWithErrorHandling(
				func(_ context.Context, e *Event) error {
					return errors.New("testing error handling decorator")
				},
			),
		},
	)

	err := fsm.Event(context.Background(), "run")
	if err == nil {
		t.Error("expected error to be returned from event")
	}
	if err.Error() != "testing error handling decorator" {
		t.Errorf("expected error to be 'testing error handling decorator', got '%s'", err.Error())
	}
}
