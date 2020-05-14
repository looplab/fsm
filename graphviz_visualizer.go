package fsm

import (
	"bytes"
	"fmt"
)

// Visualize outputs a visualization of a FSM in Graphviz format.
func Visualize(fsm *FSM) string {
	var buf bytes.Buffer

	// we sort the key alphabetically to have a reproducible graph output
	sortedEKeys := getSortedTransitionKeys(fsm.transitions)
	sortedStateKeys, _ := getSortedStates(fsm.transitions)

	writeHeaderLine(&buf)
	writeTransitions(&buf, fsm.current, sortedEKeys, fsm.transitions)
	writeStates(&buf, sortedStateKeys)
	writeFooter(&buf)

	return buf.String()
}

func writeHeaderLine(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf(`digraph fsm {`))
	buf.WriteString("\n")
}

func writeTransitions(buf *bytes.Buffer, current string, sortedEKeys []eKey, transitions map[eKey]string) {
	// make sure the current state is at top
	for _, k := range sortedEKeys {
		if k.src == current {
			v := transitions[k]
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}
	for _, k := range sortedEKeys {
		if k.src != current {
			v := transitions[k]
			buf.WriteString(fmt.Sprintf(`    "%s" -> "%s" [ label = "%s" ];`, k.src, v, k.event))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\n")
}

func writeStates(buf *bytes.Buffer, sortedStateKeys []string) {
	for _, k := range sortedStateKeys {
		buf.WriteString(fmt.Sprintf(`    "%s";`, k))
		buf.WriteString("\n")
	}
}

func writeFooter(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintln("}"))
}
