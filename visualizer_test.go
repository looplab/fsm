package fsm

import (
	"context"
	"sync"
	"testing"
)

func TestVisualizeRaceCondition(t *testing.T) {
	fsm := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
		},
		Callbacks{},
	)

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			_ = fsm.Event(context.Background(), "open")
			_ = fsm.Event(context.Background(), "close")
		}
	}()

	for i := 0; i < 100; i++ {
		_ = Visualize(fsm)
		if _, err := VisualizeForMermaidWithGraphType(fsm, StateDiagram); err != nil {
			t.Errorf("visualization failed %v", err)
		}
		if _, err := VisualizeForMermaidWithGraphType(fsm, FlowChart); err != nil {
			t.Errorf("visualization failed %v", err)
		}
	}
	close(done)
	wg.Wait()
}
