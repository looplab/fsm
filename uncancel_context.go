package fsm

import (
	"context"
	"time"
)

type uncancel struct {
	context.Context
}

func (*uncancel) Deadline() (deadline time.Time, ok bool) { return }
func (*uncancel) Done() <-chan struct{}                   { return nil }
func (*uncancel) Err() error                              { return nil }

// uncancelContext returns a context which ignores the cancellation of the parent and only keeps the values.
// Also returns a new cancel function.
// This is useful to keep a background task running while the initial request is finished.
func uncancelContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithCancel(&uncancel{ctx})
}
