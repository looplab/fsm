# FSM

[![Build Status](https://drone.io/github.com/looplab/fsm/status.png)](https://drone.io/github.com/looplab/fsm/latest)

## Summary

FSM is a finite state machine for Go.

It is heavily based on two FSM implementations:

Javascript Finite State Machine
https://github.com/jakesgordon/javascript-state-machine

Fysom for Python
https://github.com/oxplot/fysom (forked at https://github.com/mriehl/fysom)

For API docs and examples see http://godoc.org/github.com/looplab/fsm

## Example

    fsm := NewFSM(
        "closed",
        Events{
            {Name: "open", Src: []string{"closed"}, Dst: "open"},
            {Name: "close", Src: []string{"open"}, Dst: "closed"},
        },
        Callbacks{},
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

## License

FSM is licensed under Apache License 2.0

http://www.apache.org/licenses/LICENSE-2.0
