package fsm

import (
	"fmt"
	"strings"
	"testing"
)

func TestMermaidOutput(t *testing.T) {
	fsmUnderTest := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
			{Name: "part-close", Src: []string{"intermediate"}, Dst: "closed"},
		},
		Callbacks{},
	)

	got, err := VisualizeForMermaidWithGraphType(fsmUnderTest, StateDiagram)
	if err != nil {
		t.Errorf("got error for visualizing with type MERMAID: %s", err)
	}
	wanted := `
stateDiagram-v2
    [*] --> closed
    closed --> open: open
    intermediate --> closed: part-close
    open --> closed: close
`
	normalizedGot := strings.ReplaceAll(got, "\n", "")
	normalizedWanted := strings.ReplaceAll(wanted, "\n", "")
	if normalizedGot != normalizedWanted {
		t.Errorf("build mermaid graph failed. \nwanted \n%s\nand got \n%s\n", wanted, got)
		fmt.Println([]byte(normalizedGot))
		fmt.Println([]byte(normalizedWanted))
	}
}

func TestMermaidFlowChartOutput(t *testing.T) {
	fsmUnderTest := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "part-open", Src: []string{"closed"}, Dst: "intermediate"},
			{Name: "part-open", Src: []string{"intermediate"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
			{Name: "part-close", Src: []string{"intermediate"}, Dst: "closed"},
		},
		Callbacks{},
	)

	got, err := VisualizeForMermaidWithGraphType(fsmUnderTest, FlowChart)
	if err != nil {
		t.Errorf("got error for visualizing with type MERMAID: %s", err)
	}
	wanted := `
graph LR
    id0[closed]
    id1[intermediate]
    id2[open]

    id0 --> |open| id2
    id0 --> |part-open| id1
    id1 --> |part-close| id0
    id1 --> |part-open| id2
    id2 --> |close| id0

    style id0 fill:#00AA00
`
	normalizedGot := strings.ReplaceAll(got, "\n", "")
	normalizedWanted := strings.ReplaceAll(wanted, "\n", "")
	if normalizedGot != normalizedWanted {
		t.Errorf("build mermaid graph failed. \nwanted \n%s\nand got \n%s\n", wanted, got)
		fmt.Println([]byte(normalizedGot))
		fmt.Println([]byte(normalizedWanted))
	}
}
