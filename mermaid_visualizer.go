package fsm

import (
	"bytes"
	"fmt"
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
func VisualizeForMermaidWithGraphType(fsm *FSM, graphType MermaidDiagramType) (string, error) {
	switch graphType {
	case FlowChart:
		return visualizeForMermaidAsFlowChart(fsm), nil
	case StateDiagram:
		return visualizeForMermaidAsStateDiagram(fsm), nil
	default:
		return "", fmt.Errorf("unknown MermaidDiagramType: %s", graphType)
	}
}

func visualizeForMermaidAsStateDiagram(fsm *FSM) string {
	var buf bytes.Buffer

	sortedTransitionKeys := getSortedTransitionKeys(fsm.transitions)

	buf.WriteString("stateDiagram-v2\n")
	buf.WriteString(fmt.Sprintln(`    [*] -->`, fsm.current))

	for _, k := range sortedTransitionKeys {
		v := fsm.transitions[k]
		buf.WriteString(fmt.Sprintf(`    %s --> %s: %s`, k.src, v, k.event))
		buf.WriteString("\n")
	}

	return buf.String()
}

// visualizeForMermaidAsFlowChart outputs a visualization of a FSM in Mermaid format (including highlighting of current state).
func visualizeForMermaidAsFlowChart(fsm *FSM) string {
	var buf bytes.Buffer

	sortedTransitionKeys := getSortedTransitionKeys(fsm.transitions)
	sortedStates, statesToIDMap := getSortedStates(fsm.transitions)

	writeFlowChartGraphType(&buf)
	writeFlowChartStates(&buf, sortedStates, statesToIDMap)
	writeFlowChartTransitions(&buf, fsm.transitions, sortedTransitionKeys, statesToIDMap)
	writeFlowChartHightlightCurrent(&buf, fsm.current, statesToIDMap)

	return buf.String()
}

func writeFlowChartGraphType(buf *bytes.Buffer) {
	buf.WriteString("graph LR\n")
}

func writeFlowChartStates(buf *bytes.Buffer, sortedStates []string, statesToIDMap map[string]string) {
	for _, state := range sortedStates {
		buf.WriteString(fmt.Sprintf(`    %s[%s]`, statesToIDMap[state], state))
		buf.WriteString("\n")
	}

	buf.WriteString("\n")
}

func writeFlowChartTransitions(buf *bytes.Buffer, transitions map[eKey]string, sortedTransitionKeys []eKey, statesToIDMap map[string]string) {
	for _, transition := range sortedTransitionKeys {
		target := transitions[transition]
		buf.WriteString(fmt.Sprintf(`    %s --> |%s| %s`, statesToIDMap[transition.src], transition.event, statesToIDMap[target]))
		buf.WriteString("\n")
	}
	buf.WriteString("\n")
}

func writeFlowChartHightlightCurrent(buf *bytes.Buffer, current string, statesToIDMap map[string]string) {
	buf.WriteString(fmt.Sprintf(`    style %s fill:%s`, statesToIDMap[current], highlightingColor))
	buf.WriteString("\n")
}
