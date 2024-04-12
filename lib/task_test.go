package fz

import (
	"fmt"
	"testing"
	"time"
)

func TestTaskFrequency(t *testing.T) {
	var tests = []struct {
		ta   Task
		want time.Duration
	}{
		{Task{FrequencySeconds: 0}, time.Duration(300) * time.Second}, // default
		{Task{FrequencySeconds: 5}, time.Duration(5) * time.Second},
		{Task{FrequencySeconds: 3600}, time.Duration(1) * time.Hour},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("Frequency: %d", tt.ta.FrequencySeconds)
		t.Run(testname, func(t *testing.T) {
			got := tt.ta.Frequency()
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
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
		{Task{Retries: 5, FrequencySeconds: 1, history: 0b11111}, time.Unix(1712882669, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, history: 0b11111}, time.Unix(1712882670, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, history: 0b11111}, time.Unix(1712882669, 0), false},
		{Task{Retries: 5, FrequencySeconds: 10, history: 0b01011, RetryFrequencySeconds: 1}, time.Unix(1712882669, 0), true}, // should be retrying
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

func TestTaskState(t *testing.T) {
	var tests = []struct {
		ta   Task
		want int
	}{
		{Task{Retries: 5, history: 0b11111}, STATE_OK},
		{Task{Retries: 5, history: 0b00000}, STATE_FAIL},
		{Task{Retries: 5, history: 0b10111}, STATE_UNKNOWN},
		{Task{Retries: 5, history: 0b01000}, STATE_UNKNOWN},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%b", tt.ta.history)
		t.Run(testname, func(t *testing.T) {
			got := tt.ta.State()
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
