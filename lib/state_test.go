package fz

import (
	"fmt"
	"testing"
)

func TestStateAppendValue(t *testing.T) {
	st := State{0, 5, 0}

	// add a true
	st.Append(true)
	got := st.History
	want := uint32(1)

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

	// add another true
	st.Append(true)
	got = st.History
	want = uint32(3)

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

	// add a false
	st.Append(false)
	got = st.History
	want = uint32(6)

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

func TestStateStatus(t *testing.T) {

	var tests = []struct {
		st   State
		want int
	}{
		{State{0, 5, 0b11111}, 1},  // up
		{State{0, 5, 0b00000}, 0},  // down
		{State{0, 5, 0b01000}, -1}, // unknown
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%b", tt.st.History)
		t.Run(testname, func(t *testing.T) {
			got := tt.st.Status()
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
