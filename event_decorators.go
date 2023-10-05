package fsm

import "context"

// CallbackWithErr is an FSM callback function that can return an error. The primary use case for this is for
// CallbackWithErr to be used as an argument to DecorateCallbackWithErrorHandling.
type CallbackWithErr func(ctx context.Context, event *Event) error

// DecorateCallbackWithErrorHandling is a decorator for FSM callbacks that will catch any errors returned by the
// callback and set them on the event's Err field. This is useful for standardizing the way any error encountered
// during an FSM callback are handled.
// Example usage:
//
//	fsm := NewFSM(
//		"start",
//		Events{
//			{Name: "run", Src: []string{"start"}, Dst: "end"},
//		},
//		Callbacks{
//			"before_event": DecorateCallbackWithErrorHandling(
//				func(_ context.Context, e *Event) error {
//					return errors.New("testing error handling decorator")
//				},
//			),
//		},
//	)
func DecorateCallbackWithErrorHandling(callback CallbackWithErr) Callback {
	return func(ctx context.Context, event *Event) {
		err := callback(ctx, event)
		if err != nil {
			event.Err = err
		}
	}
}
