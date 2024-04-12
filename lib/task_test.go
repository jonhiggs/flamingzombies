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
		{Task{Retries: 5, FrequencySeconds: 10, history: 0b01011, RetryFrequencySeconds: 2}, time.Unix(1712882668, 0), true},
		{Task{Retries: 5, FrequencySeconds: 10, history: 0b01011, RetryFrequencySeconds: 2}, time.Unix(1712882669, 0), false},
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

func TestRecordStatus(t *testing.T) {
	// default starting state
	ta := Task{
		Retries:      3,
		history:      0b10,
		lastState:    STATE_UNKNOWN,
		stateChanged: false,
	}

	t.Run("initial:state", func(t *testing.T) {
		got := ta.State()
		want := STATE_UNKNOWN

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("initial:history", func(t *testing.T) {
		got := ta.history
		want := uint32(0b10)

		if got != want {
			t.Errorf("got %b, want %b", got, want)
		}
	})

	t.Run("initial:state_changed", func(t *testing.T) {
		got := ta.stateChanged
		want := false

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// record first measurement
	ta.RecordStatus(false)

	t.Run("first:state", func(t *testing.T) {
		got := ta.State()
		want := STATE_UNKNOWN

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("first:history", func(t *testing.T) {
		got := ta.history
		want := uint32(0b100)

		if got != want {
			t.Errorf("got %b, want %b", got, want)
		}
	})

	t.Run("first:state_changed", func(t *testing.T) {
		got := ta.stateChanged
		want := false

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// record second measurement
	ta.RecordStatus(false)

	t.Run("second:state", func(t *testing.T) {
		got := ta.State()
		want := STATE_FAIL

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("second:history", func(t *testing.T) {
		got := ta.history
		want := uint32(0b1000)

		if got != want {
			t.Errorf("got %b, want %b", got, want)
		}
	})

	t.Run("second:state_changed", func(t *testing.T) {
		got := ta.stateChanged
		want := true

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// record third measurement
	// flip to STATE_OK
	ta.RecordStatus(true)

	t.Run("third:state", func(t *testing.T) {
		got := ta.State()
		want := STATE_UNKNOWN

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("third:history", func(t *testing.T) {
		got := ta.history
		want := uint32(0b10001)

		if got != want {
			t.Errorf("got %b, want %b", got, want)
		}
	})

	t.Run("third:state_changed", func(t *testing.T) {
		got := ta.stateChanged
		want := false

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// record forth measurement
	ta.RecordStatus(true)

	t.Run("forth:state", func(t *testing.T) {
		got := ta.State()
		want := STATE_UNKNOWN

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("forth:history", func(t *testing.T) {
		got := ta.history
		want := uint32(0b100011)

		if got != want {
			t.Errorf("got %b, want %b", got, want)
		}
	})

	t.Run("forth:state_changed", func(t *testing.T) {
		got := ta.stateChanged
		want := false

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	// record fifth measurement
	ta.RecordStatus(true)

	t.Run("fifth:state", func(t *testing.T) {
		got := ta.State()
		want := STATE_OK

		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("fifth:history", func(t *testing.T) {
		got := ta.history
		want := uint32(0b1000111)

		if got != want {
			t.Errorf("got %b, want %b", got, want)
		}
	})

	t.Run("fifth:state_changed", func(t *testing.T) {
		got := ta.stateChanged
		want := true

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
