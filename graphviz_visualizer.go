package fsm

import (
	"bytes"
	"fmt"
)

// Visualize outputs a visualization of a FSM in Graphviz format.
func Visualize(fsm *FSM) string {
	var buf bytes.Buffer

	transitions := fsm.spec.transitionTable()
	current := fsm.Current()

	// we sort the key alphabetically to have a reproducible graph output
	sortedEKeys := getSortedTransitionKeys(transitions)
	sortedStateKeys, _ := getSortedStates(transitions)

	writeHeaderLine(&buf)
	writeTransitions(&buf, sortedEKeys, transitions)
	writeStates(&buf, current, sortedStateKeys)
	writeFooter(&buf)

	return buf.String()
}

func writeHeaderLine(buf *bytes.Buffer) {
	buf.WriteString(`digraph fsm {`)
	buf.WriteString("\n")
}

func writeTransitions(buf *bytes.Buffer, sortedEKeys []eKey, transitions map[eKey]string) {
	for _, k := range sortedEKeys {
		v := transitions[k]
		buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
		buf.WriteString("\n")
	}

	buf.WriteString("\n")
}

func writeStates(buf *bytes.Buffer, current string, sortedStateKeys []string) {
	for _, k := range sortedStateKeys {
		if k == current {
			buf.WriteString(fmt.Sprintf(`    "%s" [color = "red"];`, k))
		} else {
			buf.WriteString(fmt.Sprintf(`    "%s";`, k))
		}
		buf.WriteString("\n")
	}
}

func writeFooter(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintln("}"))
}
