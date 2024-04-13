package fz

import (
	"testing"
)

var testTask = Task{
	Name:        "flappy",
	ErrorBody:   "flappy has entered an error state",
	RecoverBody: "flappy has recovered",
	Retries:     3,
}
var testNotifier = Notifier{Name: "testing"}

func TestNotificationSubject(t *testing.T) {
	n := Notification{&testNotifier, &testTask}

	t.Run("when_ok", func(t *testing.T) {
		testTask.history = 0b111
		want := testTask.RecoverBody
		got := n.body()

		if got != want {
			t.Errorf("Expect '%s' but got '%s'", want, got)
		}
	})

	t.Run("when_fail", func(t *testing.T) {
		testTask.history = 0b000
		want := testTask.ErrorBody
		got := n.body()

		if got != want {
			t.Errorf("Expect '%s' but got '%s'", want, got)
		}
	})

	t.Run("when_unknown", func(t *testing.T) {
		testTask.history = 0b101
		want := "The task flappy is in an unknown state"
		got := n.body()

		if got != want {
			t.Errorf("Expect '%s' but got '%s'", want, got)
		}
	})
}
