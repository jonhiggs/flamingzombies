package fz

import (
	"fmt"
	"testing"
	"time"
)

var testTask = Task{
	Name:        "flappy",
	ErrorBody:   "flappy has entered an error state",
	RecoverBody: "flappy has recovered",
	Retries:     3,
}
var testNotifier = Notifier{Name: "testing"}

func TestNotificationSubject(t *testing.T) {
	n := Notification{
		Notifier:  &testNotifier,
		Task:      &testTask,
		Duration:  time.Second * 1,
		Timestamp: time.Now(),
	}

	t.Run("when_ok", func(t *testing.T) {
		testTask.History = 0b111
		testTask.HistoryMask = 0b111
		want := testTask.RecoverBody
		got := n.body()

		if got != want {
			t.Errorf("Expect '%s' but got '%s'", want, got)
		}
	})

	t.Run("when_fail", func(t *testing.T) {
		testTask.History = 0b000
		testTask.HistoryMask = 0b111
		want := testTask.ErrorBody
		got := n.body()

		if got != want {
			t.Errorf("Expect '%s' but got '%s'", want, got)
		}
	})

	t.Run("when_unknown", func(t *testing.T) {
		testTask.History = 0b101
		testTask.HistoryMask = 0b111
		want := "The task flappy is in an unknown state"
		got := n.body()

		if got != want {
			t.Errorf("Expect '%s' but got '%s'", want, got)
		}
	})
}

func TestNotificationEnvironment(t *testing.T) {
	cfg.Defaults.Envs = []string{"MAIL_NAME=test@example"}
	n := Notification{
		Notifier:  &testNotifier,
		Task:      &testTask,
		Duration:  time.Second * 1,
		Timestamp: time.Unix(1735517669, 0),
	}

	t.Run("env", func(t *testing.T) {

		want := []string{
			"MSG=",
			"TASK_DURATION_MS=1000",
			"TASK_EPOCH=1735517669",
			"TASK_LAST_STATE=unknown",
			"TASK_NAME=flappy",
			"TASK_PRIORITY=0",
			"TASK_STATE=unknown",
			"TASK_TIMEOUT_MS=0",
			"MAIL_NAME=test@example",
		}
		got := n.Environment(n.Task)

		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("Expect '%v' but got '%v'", want, got)
		}
	})

	// put the state back
	cfg.Defaults.Envs = []string{}
}

func TestErrorNotificationEnvironment(t *testing.T) {
	cfg.Defaults.Envs = []string{"MAIL_NAME=test@example"}
	n := ErrorNotification{
		Notifier: &testNotifier,
		Error:    fmt.Errorf("this is an error"),
	}

	t.Run("env", func(t *testing.T) {

		want := []string{
			"MSG=this is an error",
			"MAIL_NAME=test@example",
		}
		got := n.Environment()

		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("Expect '%v' but got '%v'", want, got)
		}
	})

	// put the state back
	cfg.Defaults.Envs = []string{}
}
