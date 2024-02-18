package fsm

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

func NewFSMFromTemplate(initial string, template string, callbacks map[string]Callback) (*FSM, error) {
	var events []EventDesc
	es := parseFSM(template, "Name", "Src", "Dst")
	data, err := json.Marshal(es)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, errors.New("init event fail,check your template")
	}
	return NewFSM(initial, events, callbacks), nil
}

func parseFSM(tpl string, action, from string, to string) []map[string]interface{} {
	reg := regexp.MustCompile(`(?m)^\s*//.*?\n|^\s*`)
	tpl = reg.ReplaceAllString(tpl, "")
	//fmt.Println(tpl)
	lines := strings.Split(tpl, "\n")
	stepMap := make(map[string]string)
	//actionMap := make(map[string]map[string]string)
	var actions []map[string]interface{}

	for _, line := range lines {
		kv := regexp.MustCompile(`\s*=\s*`).Split(line, -1)
		if len(kv) == 2 {
			stepMap[kv[0]] = kv[1]
		}
	}
	for _, line := range lines {
		matchers := regexp.MustCompile(`(?P<action>[\s\S]+)\s*[:：]\s*(?P<from>[\s\S]+?)\s*(?:->|→|—》)\s*(?P<to>[\s\S]+)`).FindStringSubmatch(line)
		if len(matchers) != 4 {
			continue
		}
		_from, ok := stepMap[matchers[2]]
		if !ok {
			_from = matchers[2]
		}
		_to, ok := stepMap[matchers[3]]
		if !ok {
			_to = matchers[3]
		}
		m := map[string]interface{}{
			action: matchers[1],
			from:   []string{_from},
			to:     _to,
		}
		actions = append(actions, m)

	}
	return actions
}

// GetAllEvents AllEvents
func (f *FSM) GetAllEvents() map[string]interface{} {
	events := make(map[string]interface{})
	for key, _ := range f.transitions {
		events[key.event] = key.event
	}
	return events
}

// GetAllStates AllStates
func (f *FSM) GetAllStates() map[string]interface{} {
	states := make(map[string]interface{})
	for key, val := range f.transitions {
		states[key.src] = key.src
		states[val] = val
	}
	return states
}

// Before get before states
func (f *FSM) Before(state string) []string {
	var states []string
	for src, dest := range f.transitions {
		if dest == state {
			states = append(states, src.src)
		}
	}
	return states
}

// After get after states
func (f *FSM) After(state string) []string {
	var states []string
	for src, dest := range f.transitions {
		if src.src == state {
			states = append(states, dest)
		}
	}
	return states
}
