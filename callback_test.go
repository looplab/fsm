package fsm

import "testing"

func TestCallbackValidate(t *testing.T) {
	tests := []struct {
		name      string
		cb        Callback[string, string]
		errString string
	}{
		{
			name:      "before_event without event",
			cb:        Callback[string, string]{When: BeforeEvent},
			errString: "before_event given but no event",
		},
		{
			name:      "before_event with state",
			cb:        Callback[string, string]{When: BeforeEvent, Event: "open", State: "closed"},
			errString: "before_event given but state closed specified",
		},
		{
			name:      "before_event with state",
			cb:        Callback[string, string]{When: BeforeAllEvents, Event: "open"},
			errString: "before_all_events given with event open",
		},

		{
			name:      "before_event without event",
			cb:        Callback[string, string]{When: EnterState},
			errString: "enter_state given but no state",
		},
		{
			name:      "before_event with state",
			cb:        Callback[string, string]{When: EnterState, Event: "open", State: "closed"},
			errString: "enter_state given but event open specified",
		},
		{
			name:      "before_event with state",
			cb:        Callback[string, string]{When: EnterAllStates, State: "closed"},
			errString: "enter_all_states given with state closed",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cb.validate()

			if tt.errString == "" && err != nil {
				t.Errorf("err:%v", err)
			}
			if tt.errString != "" && err == nil {
				t.Errorf("errstring:%s but err is nil", tt.errString)
			}

			if tt.errString != "" && err.Error() != tt.errString {
				t.Errorf("transition failed %v", err)
			}
		})
	}

}
