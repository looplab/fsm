package fsm

import (
	"bytes"
	"fmt"
	"sort"
)

// VisualizeType the type of the visualization
type VisualizeType string

const highlightingColor = "#00AA00"

const (
	// GRAPHVIZ the type for graphviz output (http://www.webgraphviz.com/)
	GRAPHVIZ VisualizeType = "graphviz"
	// MERMAID the type for mermaid output (https://mermaid-js.github.io/mermaid-live-editor/) in the stateDiagram form
	MERMAID VisualizeType = "mermaid"
)

// VisualizeWithType outputs a visualization of a FSM in the desired format.
// If the type is not given it defaults to GRAPHVIZ
func VisualizeWithType(fsm *FSM, visualizeType VisualizeType) (string, error) {
	switch visualizeType {
	case GRAPHVIZ:
		return Visualize(fsm), nil
	case MERMAID:
		return VisualizeForMermaidWithGraphType(fsm, StateDiagram)
	default:
		return "", fmt.Errorf("unknown VisualizeType: %s", visualizeType)
	}
}

// MermaidDiagramType the type of the mermaid diagram type
type MermaidDiagramType string

const (
	// FlowChart the diagram type for output in flowchart style (https://mermaid-js.github.io/mermaid/#/flowchart) (including current state)
	FlowChart MermaidDiagramType = "flowChart"
	// StateDiagram the diagram type for output in stateDiagram style (https://mermaid-js.github.io/mermaid/#/stateDiagram)
	StateDiagram MermaidDiagramType = "stateDiagram"
)

// VisualizeForMermaidWithGraphType outputs a visualization of a FSM in Mermaid format as specified by the graphType.
func VisualizeForMermaidWithGraphType(fsm *FSM, graphType MermaidDiagramType) (string, error){
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

	buf.WriteString("stateDiagram\n")
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

	buf.WriteString("graph LR\n")

	for _, state := range sortedStates{
		buf.WriteString(fmt.Sprintf(`    %s[%s]`,statesToIDMap[state],state))
		buf.WriteString("\n")
	}

	buf.WriteString("\n")

	for _, transition := range sortedTransitionKeys {
		target := fsm.transitions[transition]
		buf.WriteString(fmt.Sprintf(`    %s --> |%s| %s`, statesToIDMap[transition.src], transition.event, statesToIDMap[target]))
		buf.WriteString("\n")
	}
	buf.WriteString("\n")

	buf.WriteString(fmt.Sprintf(`    style %s fill:%s`, statesToIDMap[fsm.current], highlightingColor))
	buf.WriteString("\n")
	
	return buf.String()
}

func getSortedTransitionKeys(transitions map[eKey]string) []eKey {
	// we sort the key alphabetically to have a reproducible graph output
	sortedTransitionKeys := make([]eKey, 0)

	for transition:= range transitions {
		sortedTransitionKeys = append(sortedTransitionKeys, transition)
	}
	sort.Slice(sortedTransitionKeys, func(i, j int) bool {
		return sortedTransitionKeys[i].src < sortedTransitionKeys[j].src
	})

	return sortedTransitionKeys		
}

func getSortedStates(transitions map[eKey]string) ([]string, map[string]string){
	statesToIDMap := make(map[string]string)
	for transition, target := range transitions {
		if _, ok := statesToIDMap[transition.src]; !ok{
			statesToIDMap[transition.src] = "";
		}
		if _, ok := statesToIDMap[target]; !ok{
			statesToIDMap[target] = "";
		}
	}

	sortedStates := make([]string, 0, len(statesToIDMap))
	for state := range statesToIDMap {
		sortedStates = append(sortedStates, state)
	}
	sort.Strings(sortedStates)

	for i, state := range sortedStates {
		statesToIDMap[state] = fmt.Sprintf("id%d", i)
	}
	return sortedStates, statesToIDMap
}

// Visualize outputs a visualization of a FSM in Graphviz format.
func Visualize(fsm *FSM) string {
	var buf bytes.Buffer

	states := make(map[string]int)

	// we sort the key alphabetically to have a reproducible graph output
	sortedEKeys := make([]eKey, 0)
	for k := range fsm.transitions {
		sortedEKeys = append(sortedEKeys, k)
	}
	sort.Slice(sortedEKeys, func(i, j int) bool {
		return sortedEKeys[i].src < sortedEKeys[j].src
	})

	buf.WriteString(fmt.Sprintf(`digraph fsm {`))
	buf.WriteString("\n")

	// make sure the initial state is at top
	for _, k := range sortedEKeys {
		v := fsm.transitions[k]
		if k.src == fsm.current {
			states[k.src]++
			states[v]++
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	for _, k := range sortedEKeys {
		v := fsm.transitions[k]
		if k.src != fsm.current {
			states[k.src]++
			states[v]++
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\n")

	sortedStateKeys := make([]string, 0)
	for k := range states {
		sortedStateKeys = append(sortedStateKeys, k)
	}
	sort.Slice(sortedStateKeys, func(i, j int) bool {
		return sortedStateKeys[i] < sortedStateKeys[j]
	})

	for _, k := range sortedStateKeys {
		buf.WriteString(fmt.Sprintf(`    "%s";`, k))
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintln("}"))

	return buf.String()
}
