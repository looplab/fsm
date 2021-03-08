[![PkgGoDev](https://pkg.go.dev/badge/github.com/looplab/fsm)](https://pkg.go.dev/github.com/looplab/fsm)
![Bulid Status](https://github.com/looplab/fsm/actions/workflows/main.yml/badge.svg)
[![Coverage Status](https://img.shields.io/coveralls/looplab/fsm.svg)](https://coveralls.io/r/looplab/fsm)
[![Go Report Card](https://goreportcard.com/badge/looplab/fsm)](https://goreportcard.com/report/looplab/fsm)

# FSM for Go

FSM is a finite state machine for Go.

It is heavily based on two FSM implementations:

- Javascript Finite State Machine, https://github.com/jakesgordon/javascript-state-machine

- Fysom for Python, https://github.com/oxplot/fysom (forked at https://github.com/mriehl/fysom)

For API docs and examples see http://godoc.org/github.com/looplab/fsm

# Basic Example

From examples/simple.go:

```go
package main

import (
    "fmt"
    "github.com/looplab/fsm"
)

func main() {
    fsm := fsm.NewFSM(
        "closed",
        fsm.Events{
            {Name: "open", Src: []string{"closed"}, Dst: "open"},
            {Name: "close", Src: []string{"open"}, Dst: "closed"},
        },
        fsm.Callbacks{},
    )

    fmt.Println(fsm.Current())

    err := fsm.Event("open")
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(fsm.Current())

    err = fsm.Event("close")
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(fsm.Current())
}
```

# Usage as a struct field

From examples/struct.go:

```go
package main

import (
    "fmt"
    "github.com/looplab/fsm"
)

type Door struct {
    To  string
    FSM *fsm.FSM
}

func NewDoor(to string) *Door {
    d := &Door{
        To: to,
    }

    d.FSM = fsm.NewFSM(
        "closed",
        fsm.Events{
            {Name: "open", Src: []string{"closed"}, Dst: "open"},
            {Name: "close", Src: []string{"open"}, Dst: "closed"},
        },
        fsm.Callbacks{
            "enter_state": func(e *fsm.Event) { d.enterState(e) },
        },
    )

    return d
}

func (d *Door) enterState(e *fsm.Event) {
    fmt.Printf("The door to %s is %s\n", d.To, e.Dst)
}

func main() {
    door := NewDoor("heaven")

    err := door.FSM.Event("open")
    if err != nil {
        fmt.Println(err)
    }

    err = door.FSM.Event("close")
    if err != nil {
        fmt.Println(err)
    }
}
```

# License

FSM is licensed under Apache License 2.0

http://www.apache.org/licenses/LICENSE-2.0
