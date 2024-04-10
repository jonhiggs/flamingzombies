package fz

import (
	"fmt"
	"testing"
)

func TestTaskState(t *testing.T) {
	var tests = []struct {
		ta   Task
		want int
	}{
		{Task{Retries: 5, History: 0b11111}, 1},  // up
		{Task{Retries: 5, History: 0b00000}, 0},  // down
		{Task{Retries: 5, History: 0b10111}, -1}, // unknown
		{Task{Retries: 5, History: 0b01000}, -1}, // unknown
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%b", tt.ta.History)
		t.Run(testname, func(t *testing.T) {
			got := tt.ta.State()
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
