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
		matchers := regexp.MustCompile(`(?P<action>[\s\S]+):(?P<from>[\s\S]+?)\s*->\s*(?P<to>[\s\S]+)`).FindStringSubmatch(line)
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
