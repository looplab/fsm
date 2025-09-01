package fsm

import (
	"bytes"
	"cmp"
	"fmt"
)

// Visualize outputs a visualization of a FSM in Graphviz format.
func Visualize[E cmp.Ordered, S cmp.Ordered](fsm *FSM[E, S]) string {
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
	buf.WriteString(`digraph fsm {`)
	buf.WriteString("\n")
}

func writeTransitions[E cmp.Ordered, S cmp.Ordered](buf *bytes.Buffer, current S, sortedEKeys []eKey[E, S], transitions map[eKey[E, S]]S) {
	// make sure the current state is at top
	for _, k := range sortedEKeys {
		if k.src == current {
			v := transitions[k]
			fmt.Fprintf(buf, `    "%v" -> "%v" [ label = "%v" ];`, k.src, v, k.event)
			buf.WriteString("\n")
		}
	}
	for _, k := range sortedEKeys {
		if k.src != current {
			v := transitions[k]
			fmt.Fprintf(buf, `    "%v" -> "%v" [ label = "%v" ];`, k.src, v, k.event)
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\n")
}

func writeStates[S cmp.Ordered](buf *bytes.Buffer, sortedStateKeys []S) {
	for _, k := range sortedStateKeys {
		fmt.Fprintf(buf, `    "%v";`, k)
		buf.WriteString("\n")
	}
}

func writeFooter(buf *bytes.Buffer) {
	fmt.Fprintln(buf, "}")
}
