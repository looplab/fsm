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
    "context"
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

    err := fsm.Event(context.Background(), "open")
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println(fsm.Current())

    err = fsm.Event(context.Background(), "close")
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
    "context"
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
            "enter_state": func(_ context.Context, e *fsm.Event) { d.enterState(e) },
        },
    )

    return d
}

func (d *Door) enterState(e *fsm.Event) {
    fmt.Printf("The door to %s is %s\n", d.To, e.Dst)
}

func main() {
    door := NewDoor("heaven")

    err := door.FSM.Event(context.Background(), "open")
    if err != nil {
        fmt.Println(err)
    }

    err = door.FSM.Event(context.Background(), "close")
    if err != nil {
        fmt.Println(err)
    }
}
```

# Shared machine definitions

Applications that keep one FSM per managed object can compile the events and
callbacks once into a `Spec` and share it between all machines, instead of
building the same transition and callback maps for every machine:

```go
package main

import (
    "context"
    "fmt"

    "github.com/looplab/fsm"
)

var doorSpec = fsm.NewSpec(
    fsm.Events{
        {Name: "open", Src: []string{"closed"}, Dst: "open"},
        {Name: "close", Src: []string{"open"}, Dst: "closed"},
    },
    fsm.Callbacks{
        "enter_state": func(_ context.Context, e *fsm.Event) {
            fmt.Println("the door is", e.Dst)
        },
    },
)

func main() {
    for i := 0; i < 3; i++ {
        door := fsm.NewFSMFromSpec("closed", doorSpec)
        if err := door.Event(context.Background(), "open"); err != nil {
            fmt.Println(err)
        }
    }
}
```

A `Spec` is immutable and holds no state of its own, so any number of machines
can use it, also concurrently. Everything that changes while a machine runs is
kept per machine.

# License

FSM is licensed under Apache License 2.0

http://www.apache.org/licenses/LICENSE-2.0
