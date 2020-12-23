package fsm

import (
	"fmt"
	"sort"
)

// VisualizeType the type of the visualization
type VisualizeType string

const (
	// GRAPHVIZ the type for graphviz output (http://www.webgraphviz.com/)
	GRAPHVIZ VisualizeType = "graphviz"
	// MERMAID the type for mermaid output (https://mermaid-js.github.io/mermaid/#/stateDiagram) in the stateDiagram form
	MERMAID VisualizeType = "mermaid"
	// MermaidStateDiagram the type for mermaid output (https://mermaid-js.github.io/mermaid/#/stateDiagram) in the stateDiagram form
	MermaidStateDiagram VisualizeType = "mermaid-state-diagram"
	// MermaidFlowChart the type for mermaid output (https://mermaid-js.github.io/mermaid/#/flowchart) in the flow chart form
	MermaidFlowChart VisualizeType = "mermaid-flow-chart"
)

// VisualizeWithType outputs a visualization of a FSM in the desired format.
// If the type is not given it defaults to GRAPHVIZ
func VisualizeWithType(fsm *FSM, visualizeType VisualizeType) (string, error) {
	switch visualizeType {
	case GRAPHVIZ:
		return Visualize(fsm), nil
	case MERMAID:
		return VisualizeForMermaidWithGraphType(fsm, StateDiagram)
	case MermaidStateDiagram:
		return VisualizeForMermaidWithGraphType(fsm, StateDiagram)
	case MermaidFlowChart:
		return VisualizeForMermaidWithGraphType(fsm, FlowChart)
	default:
		return "", fmt.Errorf("unknown VisualizeType: %s", visualizeType)
	}
}

func getSortedTransitionKeys(transitions map[eKey]string) []eKey {
	// we sort the key alphabetically to have a reproducible graph output
	sortedTransitionKeys := make([]eKey, 0)

	for transition := range transitions {
		sortedTransitionKeys = append(sortedTransitionKeys, transition)
	}
	sort.Slice(sortedTransitionKeys, func(i, j int) bool {
		if sortedTransitionKeys[i].src == sortedTransitionKeys[j].src {
			return sortedTransitionKeys[i].event < sortedTransitionKeys[j].event
		}
		return sortedTransitionKeys[i].src < sortedTransitionKeys[j].src
	})

	return sortedTransitionKeys
}

func getSortedStates(transitions map[eKey]string) ([]string, map[string]string) {
	statesToIDMap := make(map[string]string)
	for transition, target := range transitions {
		if _, ok := statesToIDMap[transition.src]; !ok {
			statesToIDMap[transition.src] = ""
		}
		if _, ok := statesToIDMap[target]; !ok {
			statesToIDMap[target] = ""
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
