package fsm

import (
	"context"
	"testing"
)

func TestUncancel(t *testing.T) {
	t.Run("create a new context", func(t *testing.T) {
		t.Run("and cancel it", func(t *testing.T) {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "key1", "value1")
			ctx, cancelFunc := context.WithCancel(ctx)
			cancelFunc()

			if ctx.Err() != context.Canceled {
				t.Errorf("expected context error 'context canceled', got %v", ctx.Err())
			}
			select {
			case <-ctx.Done():
			default:
				t.Error("expected context to be done but it wasn't")
			}

			t.Run("and uncancel it", func(t *testing.T) {
				ctx, newCancelFunc := uncancelContext(ctx)
				if ctx.Err() != nil {
					t.Errorf("expected context error to be nil, got %v", ctx.Err())
				}
				select {
				case <-ctx.Done():
					t.Fail()
				default:
				}

				t.Run("now it should still contain the values", func(t *testing.T) {
					if ctx.Value("key1") != "value1" {
						t.Errorf("expected context value of key 'key1' to be 'value1', got %v", ctx.Value("key1"))
					}
				})
				t.Run("and cancel the child", func(t *testing.T) {
					newCancelFunc()
					if ctx.Err() != context.Canceled {
						t.Errorf("expected context error 'context canceled', got %v", ctx.Err())
					}
					select {
					case <-ctx.Done():
					default:
						t.Error("expected context to be done but it wasn't")
					}
				})
			})
		})
		t.Run("and uncancel it", func(t *testing.T) {
			ctx := context.Background()
			parent := ctx
			ctx, newCancelFunc := uncancelContext(ctx)
			if ctx.Err() != nil {
				t.Errorf("expected context error to be nil, got %v", ctx.Err())
			}
			select {
			case <-ctx.Done():
				t.Fail()
			default:
			}

			t.Run("and cancel the child", func(t *testing.T) {
				newCancelFunc()
				if ctx.Err() != context.Canceled {
					t.Errorf("expected context error 'context canceled', got %v", ctx.Err())
				}
				select {
				case <-ctx.Done():
				default:
					t.Error("expected context to be done but it wasn't")
				}

				t.Run("and ensure the parent is not affected", func(t *testing.T) {
					if parent.Err() != nil {
						t.Errorf("expected parent context error to be nil, got %v", ctx.Err())
					}
					select {
					case <-parent.Done():
						t.Fail()
					default:
					}
				})
			})
		})
	})
}
