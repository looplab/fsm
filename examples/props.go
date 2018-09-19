// +build ignore

package main

import (
	"fmt"
	"github.com/fari-99/fsm"
)

func main() {
	fsm := fsm.NewFSM(
		"closed",
		fsm.Events{
			{
				Name: "open",
				Src: []string{"closed"},
				Dst: "open",
		},
			{
				Name: "close",
				Src: []string{"open"},
				Dst: "closed"},
		},
		fsm.Callbacks{},
	)

	fmt.Println(fsm.Current())
}
