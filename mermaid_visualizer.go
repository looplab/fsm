package fsm

import (
	"bytes"
	"fmt"

	"golang.org/x/exp/constraints"
)

const highlightingColor = "#00AA00"

// MermaidDiagramType the type of the mermaid diagram type
type MermaidDiagramType string

const (
	// FlowChart the diagram type for output in flowchart style (https://mermaid-js.github.io/mermaid/#/flowchart) (including current state)
	FlowChart MermaidDiagramType = "flowChart"
	// StateDiagram the diagram type for output in stateDiagram style (https://mermaid-js.github.io/mermaid/#/stateDiagram)
	StateDiagram MermaidDiagramType = "stateDiagram"
)

// VisualizeForMermaidWithGraphType outputs a visualization of a FSM in Mermaid format as specified by the graphType.
func VisualizeForMermaidWithGraphType[E constraints.Ordered, S constraints.Ordered](fsm *FSM[E, S], graphType MermaidDiagramType) (string, error) {
	switch graphType {
	case FlowChart:
		return visualizeForMermaidAsFlowChart(fsm), nil
	case StateDiagram:
		return visualizeForMermaidAsStateDiagram(fsm), nil
	default:
		return "", fmt.Errorf("unknown MermaidDiagramType: %s", graphType)
	}
}

func visualizeForMermaidAsStateDiagram[E constraints.Ordered, S constraints.Ordered](fsm *FSM[E, S]) string {
	var buf bytes.Buffer

	sortedTransitionKeys := getSortedTransitionKeys(fsm.transitions)

	buf.WriteString("stateDiagram-v2\n")
	buf.WriteString(fmt.Sprintln(`    [*] -->`, fsm.current))

	for _, k := range sortedTransitionKeys {
		v := fsm.transitions[k]
		buf.WriteString(fmt.Sprintf(`    %v --> %v: %v`, k.src, v, k.event))
		buf.WriteString("\n")
	}

	return buf.String()
}

// visualizeForMermaidAsFlowChart outputs a visualization of a FSM in Mermaid format (including highlighting of current state).
func visualizeForMermaidAsFlowChart[E constraints.Ordered, S constraints.Ordered](fsm *FSM[E, S]) string {
	var buf bytes.Buffer

	sortedTransitionKeys := getSortedTransitionKeys(fsm.transitions)
	sortedStates, statesToIDMap := getSortedStates(fsm.transitions)

	writeFlowChartGraphType(&buf)
	writeFlowChartStates(&buf, sortedStates, statesToIDMap)
	writeFlowChartTransitions(&buf, fsm.transitions, sortedTransitionKeys, statesToIDMap)
	writeFlowChartHighlightCurrent(&buf, fsm.current, statesToIDMap)

	return buf.String()
}

func writeFlowChartGraphType(buf *bytes.Buffer) {
	buf.WriteString("graph LR\n")
}

func writeFlowChartStates[S constraints.Ordered](buf *bytes.Buffer, sortedStates []S, statesToIDMap map[S]string) {
	for _, state := range sortedStates {
		buf.WriteString(fmt.Sprintf(`    %s[%v]`, statesToIDMap[state], state))
		buf.WriteString("\n")
	}

	buf.WriteString("\n")
}

func writeFlowChartTransitions[E constraints.Ordered, S constraints.Ordered](buf *bytes.Buffer, transitions map[eKey[E, S]]S, sortedTransitionKeys []eKey[E, S], statesToIDMap map[S]string) {
	for _, transition := range sortedTransitionKeys {
		target := transitions[transition]
		buf.WriteString(fmt.Sprintf(`    %s --> |%v| %s`, statesToIDMap[transition.src], transition.event, statesToIDMap[target]))
		buf.WriteString("\n")
	}
	buf.WriteString("\n")
}

func writeFlowChartHighlightCurrent[S constraints.Ordered](buf *bytes.Buffer, current S, statesToIDMap map[S]string) {
	buf.WriteString(fmt.Sprintf(`    style %s fill:%s`, statesToIDMap[current], highlightingColor))
	buf.WriteString("\n")
}
