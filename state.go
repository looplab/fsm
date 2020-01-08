package fsm

import "fmt"

const (
	TerminalStateClassification = "Terminal"
	NonTerminalStateClassification = "Non-terminal"
)

func NewTerminalState(name string) State{
	return State{
		name: name,
		classification: TerminalStateClassification,
	}
}

func NewNonTerminalState(name string) State{
	return State{
		name: name,
		classification: NonTerminalStateClassification,
	}
}

type State struct {
	name string
	classification string
}

func (s State) Equal(another State) bool {
	return s.name == another.GetName() && s.GetType() == s.GetType()
}

func (s State) IsNamed(name string) bool {
	return s.name == name
}

func (s State) IsTerminal() bool {
	return s.classification == TerminalStateClassification
}

func (s State) IsNonTerminal() bool {
	return s.classification == NonTerminalStateClassification
}

func (s State) GetName() string {
	return s.name
}

func (s State) GetType() string {
	return s.classification
}

func (s State) String() string{
	return fmt.Sprintf("(Name: %v) (Type: %v)", s.name, s.classification)
}

