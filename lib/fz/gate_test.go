package fz

import (
	"fmt"
	"testing"
)

func TestGateEnvironment(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		gate := Gate{
			Name: "zero",
		}
		task := Task{
			Name:             "test",
			Command:          "true",
			FrequencySeconds: 60,
			Priority:         3,
		}

		got := gate.Environment(task)
		want := []string{
			"GATE_NAME=zero",
			"GATE_TIMEOUT=1",
			"TASK_COMMAND=true",
			"TASK_FREQUENCY=60",
			"TASK_HISTORY=0",
			"TASK_HISTORY_MASK=0",
			"TASK_LAST_FAIL=0",
			"TASK_LAST_OK=0",
			"TASK_LAST_STATE=ok",
			"TASK_NAME=test",
			"TASK_PRIORITY=3",
			"TASK_STATE=fail",
			"TASK_STATE_CHANGED=false",
			"TASK_TIMEOUT=0",
		}

		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("with env", func(t *testing.T) {
		gate := Gate{
			Name: "zero",
			Envs: []string{"GENV=from_gate"},
		}
		task := Task{
			Name:             "test",
			Command:          "true",
			Envs:             []string{"TENV=from_task"},
			FrequencySeconds: 60,
			Priority:         3,
		}

		got := gate.Environment(task)
		want := []string{
			"GATE_NAME=zero",
			"GATE_TIMEOUT=1",
			"GENV=from_gate",
			"TASK_COMMAND=true",
			"TASK_FREQUENCY=60",
			"TASK_HISTORY=0",
			"TASK_HISTORY_MASK=0",
			"TASK_LAST_FAIL=0",
			"TASK_LAST_OK=0",
			"TASK_LAST_STATE=ok",
			"TASK_NAME=test",
			"TASK_PRIORITY=3",
			"TASK_STATE=fail",
			"TASK_STATE_CHANGED=false",
			"TASK_TIMEOUT=0",
			"TENV=from_task",
		}

		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
