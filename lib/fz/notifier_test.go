package fz

import (
	"testing"
	"time"
)

var notifiers = []Notifier{
	Notifier{
		Name:           "zero",
		TimeoutSeconds: 3,
		GateSets: [][]string{
			[]string{"gate_zero"},
		},
	},
}

func TestNotifierTimeout(t *testing.T) {
	got := notifiers[0].Timeout()
	want := time.Second * 3

	if got != want {
		t.Errorf("length: got %d, want %d", got, want)
	}
}
