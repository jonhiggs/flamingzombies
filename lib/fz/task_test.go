package fz

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func init() {
	StartLogger("info")
}

func TestTaskValidate(t *testing.T) {
	var tests = []struct {
		task Task
		want error
	}{
		{
			Task{
				Name:    "name with spaces",
				Command: "true",
			},
			ErrInvalidName,
		},
		{
			Task{
				Name:                  "fast_retry",
				Command:               "true",
				FrequencySeconds:      5,
				RetryFrequencySeconds: 0,
			},
			ErrLessThan1,
		},
		{
			Task{
				Name:                  "too_frequent",
				Command:               "true",
				FrequencySeconds:      0,
				RetryFrequencySeconds: 5,
			},
			ErrLessThan1,
		},
		{
			Task{
				Name:                  "zero_retry_frequency",
				Command:               "true",
				FrequencySeconds:      60,
				RetryFrequencySeconds: 0,
			},
			ErrLessThan1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.task.Name, func(t *testing.T) {
			got := errors.Unwrap(tt.task.Validate())
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskState(t *testing.T) {
	tests := []struct {
		name string
		ta   Task
		want State
	}{
		{"no_measurements", Task{Retries: 3, History: 0b0, HistoryMask: 0b0}, STATE_UNKNOWN},
		{"few_measurements", Task{Retries: 3, History: 0b1, HistoryMask: 0b1}, STATE_UNKNOWN},
		{"ok", Task{Retries: 3, History: 0b111, HistoryMask: 0b111}, STATE_OK},
		{"fail", Task{Retries: 3, History: 0b000, HistoryMask: 0b111}, STATE_FAIL},
		{"to_ok", Task{Retries: 3, History: 0b000111, HistoryMask: 0b111111}, STATE_OK},
		{"to_fail", Task{Retries: 3, History: 0b111000, HistoryMask: 0b111111}, STATE_FAIL},
		{"to_unknown", Task{Retries: 3, History: 0b11100, HistoryMask: 0b11111}, STATE_UNKNOWN},
		{"big_test", Task{Retries: 3, History: 0b11000000000000000000001111111111, HistoryMask: 0b11111111111111111111111111111111}, STATE_OK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ta.State()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTaskLastState(t *testing.T) {
	tests := []struct {
		name string
		ta   Task
		want State
	}{
		{"nothing", Task{Retries: 3, History: 0b0, HistoryMask: 0b0}, STATE_UNKNOWN},
		{"few", Task{Retries: 3, History: 0b11, HistoryMask: 0b11}, STATE_UNKNOWN},
		{"one_measure", Task{Retries: 3, History: 0b111, HistoryMask: 0b111}, STATE_UNKNOWN},
		{"one_and_half_measures", Task{Retries: 3, History: 0b11111, HistoryMask: 0b11111}, STATE_UNKNOWN},
		{"two_measures", Task{Retries: 3, History: 0b111111, HistoryMask: 0b111111}, STATE_OK},
		{"ok_flap_fail", Task{Retries: 3, History: 0b11101010101000, HistoryMask: 0b11111111111111}, STATE_OK},
		{"fail_flap_ok", Task{Retries: 3, History: 0b00010101010111, HistoryMask: 0b11111111111111}, STATE_FAIL},
		{"fail_flap_fail", Task{Retries: 3, History: 0b00010101101000, HistoryMask: 0b11111111111111}, STATE_FAIL},
		{"ok_fail_flap_fail", Task{Retries: 3, History: 0b11100010101101000, HistoryMask: 0b11111111111111111}, STATE_FAIL},
		{"big_test", Task{Retries: 3, History: 0b11000000000000000000001111111111, HistoryMask: 0b11111111111111111111111111111111}, STATE_OK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ta.LastState()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTaskStateChanged(t *testing.T) {
	tests := []struct {
		name string
		ta   Task
		want bool
	}{
		{"nothing", Task{Retries: 3, History: 0b0, HistoryMask: 0b0}, false},
		{"few", Task{Retries: 3, History: 0b11, HistoryMask: 0b11}, false},
		{"one_measure", Task{Retries: 3, History: 0b111, HistoryMask: 0b111}, false},
		{"one_and_half_measures", Task{Retries: 3, History: 0b11111, HistoryMask: 0b11111}, false},
		{"two_measures", Task{Retries: 3, History: 0b111111, HistoryMask: 0b111111}, false},
		{"ok_flap_fail", Task{Retries: 3, History: 0b11101010101000, HistoryMask: 0b11111111111111}, true},
		{"fail_flap_ok", Task{Retries: 3, History: 0b00010101010111, HistoryMask: 0b11111111111111}, true},
		{"fail_flap_fail", Task{Retries: 3, History: 0b00010101101000, HistoryMask: 0b11111111111111}, false},
		{"ok_fail_flap_fail", Task{Retries: 3, History: 0b11100010101101000, HistoryMask: 0b11111111111111111}, false},
		{"big_test", Task{Retries: 3, History: 0b11000000000000000000001111111111, HistoryMask: 0b11111111111111111111111111111111}, false},
		{"unexpected_failure", Task{Retries: 5, History: 0b10000001111111111, HistoryMask: 0b10000001111111111}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ta.StateChanged()
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskReady(t *testing.T) {
	var tests = []struct {
		ta   Task
		ts   time.Time
		want bool
	}{
		{Task{Retries: 5, FrequencySeconds: 1, History: 0b11111, HistoryMask: 0b11111}, time.Unix(1712882669, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, History: 0b11111, HistoryMask: 0b11111}, time.Unix(1712882670, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, History: 0b11111, HistoryMask: 0b11111}, time.Unix(1712882669, 0), false},
		{Task{Retries: 5, FrequencySeconds: 10, History: 0b01011, HistoryMask: 0b11111, RetryFrequencySeconds: 2}, time.Unix(1712882668, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, History: 0b01011, HistoryMask: 0b11111, RetryFrequencySeconds: 2}, time.Unix(1712882669, 0), false},

		// an unknown task should test at the frequency of RetryFrequencySeconds
		{Task{Retries: 5, FrequencySeconds: 10, RetryFrequencySeconds: 1, History: 0b00001, HistoryMask: 0b00001}, time.Unix(1712882670, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, RetryFrequencySeconds: 1, History: 0b00001, HistoryMask: 0b00001}, time.Unix(1712882671, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, RetryFrequencySeconds: 10, History: 0b00001, HistoryMask: 0b00001}, time.Unix(1712882671, 0), false},

		// a passing task should test at FrequencySeconds
		{Task{Retries: 5, FrequencySeconds: 60, RetryFrequencySeconds: 10, History: 0b11111, HistoryMask: 0b11111}, time.Unix(1712882670, 0), false},
		{Task{Retries: 5, FrequencySeconds: 60, RetryFrequencySeconds: 10, History: 0b11111, HistoryMask: 0b11111}, time.Unix(1712882700, 0), true},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("ts:%d freq:%d", tt.ts.Unix(), tt.ta.FrequencySeconds)
		t.Run(testname, func(t *testing.T) {
			got := tt.ta.Ready(tt.ts)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecordStatus(t *testing.T) {
	type testTask struct {
		preRecord        bool
		wantHistory      uint32
		wantHistoryMask  uint32
		wantState        State
		wantLastState    State
		wantStateChanged bool
	}

	var testTaskSequence = []testTask{
		testTask{false, 0b0, 0b000001, STATE_UNKNOWN, STATE_UNKNOWN, false},
		testTask{false, 0b0, 0b000011, STATE_UNKNOWN, STATE_UNKNOWN, false},
		testTask{false, 0b0, 0b000111, STATE_FAIL, STATE_UNKNOWN, false},
		testTask{false, 0b0, 0b001111, STATE_FAIL, STATE_UNKNOWN, false},
		testTask{false, 0b0, 0b011111, STATE_FAIL, STATE_UNKNOWN, false},
		testTask{false, 0b0, 0b111111, STATE_FAIL, STATE_FAIL, false},
		testTask{true, 0b0000001, 0b000001111111, STATE_UNKNOWN, STATE_FAIL, false},
		testTask{true, 0b0000011, 0b000011111111, STATE_UNKNOWN, STATE_FAIL, false},
		testTask{true, 0b0000111, 0b000111111111, STATE_OK, STATE_FAIL, true},
		testTask{true, 0b0001111, 0b001111111111, STATE_OK, STATE_FAIL, false},
		testTask{true, 0b0011111, 0b011111111111, STATE_OK, STATE_FAIL, false},
	}

	// default starting state
	ta := Task{Retries: 3}

	for i, tt := range testTaskSequence {
		ta.RecordStatus(tt.preRecord)

		t.Run(fmt.Sprintf("seq_%d:history", i), func(t *testing.T) {
			got := ta.History

			if got != tt.wantHistory {
				t.Errorf("got %b, want %b", got, tt.wantHistory)
			}
		})

		t.Run(fmt.Sprintf("seq_%d:measurements", i), func(t *testing.T) {
			got := ta.HistoryMask

			if got != tt.wantHistoryMask {
				t.Errorf("got %b, want %b", got, tt.wantHistoryMask)
			}
		})

		t.Run(fmt.Sprintf("seq_%d:state", i), func(t *testing.T) {
			got := ta.State()

			if got != tt.wantState {
				t.Errorf("got %s, want %s", got, tt.wantState)
			}
		})

		t.Run(fmt.Sprintf("seq_%d:last_state", i), func(t *testing.T) {
			got := ta.LastState()

			if got != tt.wantLastState {
				t.Errorf("got %s, want %s", got, tt.wantLastState)
			}
		})

		t.Run(fmt.Sprintf("seq_%d:state_changed", i), func(t *testing.T) {
			got := ta.StateChanged()

			if got != tt.wantStateChanged {
				t.Errorf("got %v, want %v", got, tt.wantStateChanged)
			}
		})
	}
}

func TestTaskEnvironment(t *testing.T) {
	tests := []struct {
		task Task
		want []string
	}{
		{
			Task{
				Command:               "true",
				Envs:                  []string{},
				FrequencySeconds:      300,
				History:               0b10,
				Name:                  "nothing",
				Priority:              3,
				Retries:               4,
				RetryFrequencySeconds: 60,
				TimeoutSeconds:        10,
			},
			[]string{
				"TASK_COMMAND=true",
				"TASK_FREQUENCY=300",
				"TASK_HISTORY=2",
				"TASK_HISTORY_MASK=0",
				"TASK_LAST_FAIL=0",
				"TASK_LAST_OK=0",
				"TASK_LAST_STATE=unknown",
				"TASK_NAME=nothing",
				"TASK_PRIORITY=3",
				"TASK_STATE=unknown",
				"TASK_STATE_CHANGED=false",
				"TASK_TIMEOUT=10",
			},
		},
		{
			Task{
				Command: "true",
				Envs: []string{
					"SNMP_COMMUNITY=public",
					"SNMP_VERSION=2c",
				},
				FrequencySeconds:      300,
				History:               0b10,
				Name:                  "with_snmp_vars",
				Priority:              3,
				Retries:               4,
				RetryFrequencySeconds: 60,
				TimeoutSeconds:        10,
			},
			[]string{
				"TASK_COMMAND=true",
				"TASK_FREQUENCY=300",
				"TASK_HISTORY=2",
				"TASK_HISTORY_MASK=0",
				"TASK_LAST_FAIL=0",
				"TASK_LAST_OK=0",
				"TASK_LAST_STATE=unknown",
				"TASK_NAME=with_snmp_vars",
				"TASK_PRIORITY=3",
				"TASK_STATE=unknown",
				"TASK_STATE_CHANGED=false",
				"TASK_TIMEOUT=10",
				"SNMP_COMMUNITY=public",
				"SNMP_VERSION=2c",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.task.Name, func(t *testing.T) {
			// we only need to check data that's valid
			if tt.task.Validate() != nil {
				t.Errorf("%s: valid got %v, want %v", tt.task.Name, tt.task.Validate(), nil)
			}

			if fmt.Sprintf("%v", tt.task.Environment()) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("%s: environment got %v, want %v", tt.task.Name, tt.task.Environment(), tt.want)
			}
		})
	}
}
