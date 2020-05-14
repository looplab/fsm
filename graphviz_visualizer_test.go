package fsm

import (
	"fmt"
	"strings"
	"testing"
)

func TestGraphvizOutput(t *testing.T) {
	fsmUnderTest := NewFSM(
		"closed",
		Events{
			{Name: "open", Src: []string{"closed"}, Dst: "open"},
			{Name: "close", Src: []string{"open"}, Dst: "closed"},
			{Name: "part-close", Src: []string{"intermediate"}, Dst: "closed"},
		},
		Callbacks{},
	)

	got := Visualize(fsmUnderTest)
	wanted := `
digraph fsm {
    "closed" -> "open" [ label = "open" ];
    "intermediate" -> "closed" [ label = "part-close" ];
    "open" -> "closed" [ label = "close" ];

    "closed";
    "intermediate";
    "open";
}`
	normalizedGot := strings.ReplaceAll(got, "\n", "")
	normalizedWanted := strings.ReplaceAll(wanted, "\n", "")
	if normalizedGot != normalizedWanted {
		t.Errorf("build graphivz graph failed. \nwanted \n%s\nand got \n%s\n", wanted, got)
		fmt.Println([]byte(normalizedGot))
		fmt.Println([]byte(normalizedWanted))
	}
}
