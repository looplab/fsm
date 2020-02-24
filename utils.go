package fsm

import (
	"bytes"
	"fmt"
)

// VisualizeType the type of the visualization
type VisualizeType string

const (
	// GRAPHVIZ the type for graphviz output (http://www.webgraphviz.com/)
	GRAPHVIZ VisualizeType = "graphviz"
	// MERMAID the type for mermaid output (https://mermaid-js.github.io/mermaid-live-editor/)
	MERMAID VisualizeType = "mermaid"
)

// VisualizeWithType outputs a visualization of a FSM in the desired format.
// If the type is not given it defaults to GRAPHVIZ
func VisualizeWithType(fsm *FSM, visualizeType VisualizeType) (string, error) {
	switch visualizeType {
	case GRAPHVIZ:
		return Visualize(fsm), nil
	case MERMAID:
		return visualizeForMermaid(fsm), nil
	default:
		return "", fmt.Errorf("unknown VisualizeType: %s", visualizeType)
	}
}

// visualizeForMermaid outputs a visualization of a FSM in Mermaid format.
func visualizeForMermaid(fsm *FSM) string {
	var buf bytes.Buffer

	states := make(map[string]int)

	buf.WriteString(fmt.Sprintf(`graph fsm`))
	buf.WriteString("\n")

	for k, v := range fsm.transitions {
		states[k.src]++
		states[v]++
		buf.WriteString(fmt.Sprintf(`    %s -->|%s| %s`, k.src, k.event, v))
		buf.WriteString("\n")
	}

	return buf.String()
}

// Visualize outputs a visualization of a FSM in Graphviz format.
func Visualize(fsm *FSM) string {
	var buf bytes.Buffer

	states := make(map[string]int)

	buf.WriteString(fmt.Sprintf(`digraph fsm {`))
	buf.WriteString("\n")

	// make sure the initial state is at top
	for k, v := range fsm.transitions {
		if k.src == fsm.current {
			states[k.src]++
			states[v]++
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	for k, v := range fsm.transitions {
		if k.src != fsm.current {
			states[k.src]++
			states[v]++
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\n")

	for k := range states {
		buf.WriteString(fmt.Sprintf(`    "%s";`, k))
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintln("}"))

	return buf.String()
}
