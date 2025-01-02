package fz

import (
	"fmt"
	"testing"
	"time"
)

var testTask = Task{
	Name:             "flappy",
	ErrorBody:        "flappy has entered an error state",
	RecoverBody:      "flappy has recovered",
	Retries:          3,
	LastNotification: time.Unix(0, 0),
	TraceID:          "123",
}
var testNotifier = Notifier{Name: "testing"}

func TestNotificationEnvironment(t *testing.T) {
	cfg.Defaults.Envs = []string{"MAIL_NAME=test@example"}
	n := TaskNotification{
		Notifier:  &testNotifier,
		Task:      &testTask,
		Duration:  time.Second * 1,
		Timestamp: time.Unix(1735517669, 0),
	}

	t.Run("env", func(t *testing.T) {

		want := []string{
			"MSG=no message recieved",
			"SUBJECT=flappy: state is unknown",
			"TASK_DESCRIPTION=no description",
			"TASK_DURATION_MS=1000",
			"TASK_EPOCH=1735517669",
			"TASK_LAST_NOTIFICATION=0",
			"TASK_LAST_STATE=unknown",
			"TASK_NAME=flappy",
			"TASK_PRIORITY=0",
			"TASK_STATE=unknown",
			"TASK_TIMEOUT_MS=0",
			"TASK_TRACE_ID=123",
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
		TraceID:  "ABC",
	}

	t.Run("env", func(t *testing.T) {

		want := []string{
			"MSG=this is an error",
			"SUBJECT=fz experienced a critical error",
			"TASK_DESCRIPTION=An unexpected error occurred",
			"TASK_TRACE_ID=ABC",
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
