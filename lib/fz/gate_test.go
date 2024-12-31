package fz

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"
)

func TestGateEnvironment(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		gate := Gate{
			Name: "zero",
		}
		task := Task{
			TraceID:          "ABC",
			Name:             "test",
			Command:          "true",
			FrequencySeconds: 60,
			Priority:         3,
		}

		got := gate.environment(&task)
		want := []string{
			"GATE_NAME=zero",
			"GATE_TIMEOUT=1",
			"TASK_TRACE_ID=ABC",
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
			TraceID:          "ABC",
			Name:             "test",
			Command:          "true",
			Envs:             []string{"TENV=from_task"},
			FrequencySeconds: 60,
			Priority:         3,
		}

		got := gate.environment(&task)
		want := []string{
			"GATE_NAME=zero",
			"GATE_TIMEOUT=1",
			"GENV=from_gate",
			"TASK_TRACE_ID=ABC",
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

func TestGateIsOpen(t *testing.T) {
	t.Run("when open", func(t *testing.T) {
		gate := Gate{
			Name:    "true",
			Command: "echo",
			Args:    []string{"-n", "hello", "world"},
		}
		task := Task{
			TraceID: "ABC",
			Name:    "test",
		}

		gotB, gotR := gate.IsOpen(&task)

		if gotB != true {
			t.Errorf("open state: got %v, want true", gotB)
		}

		if gotR.ExitCode != 0 {
			t.Errorf("exit code: got %d, want %d", gotR.ExitCode, 0)
		}

		if string(gotR.StdoutBytes) != "hello world" {
			t.Errorf("stdout: got %s, want %s", string(gotR.StdoutBytes), "hello world")
		}

		if gotR.Duration == 0 {
			t.Errorf("duration: got %d, want something greater than 0", gotR.Duration)
		}

		if gotR.Err != nil {
			t.Errorf("error: got %s, want nil", gotR.Err)
		}
	})

	t.Run("when closed", func(t *testing.T) {
		gate := Gate{
			Name:    "false",
			Command: "false",
		}
		task := Task{
			TraceID: "ABC",
			Name:    "test",
		}

		gotB, gotR := gate.IsOpen(&task)

		if gotB != false {
			t.Errorf("got %v, want false", gotB)
		}

		if gotR.ExitCode == 0 {
			t.Errorf("exit code: got %d, want something other than 0", gotR.ExitCode)
		}

		if string(gotR.StdoutBytes) != "" {
			t.Errorf("stdout: got %s, want %s", string(gotR.StdoutBytes), "")
		}

		if gotR.Duration == 0 {
			t.Errorf("duration: got %d, want something greater than 0", gotR.Duration)
		}

		if gotR.Err != nil {
			t.Errorf("error: got %s, want nil", gotR.Err)
		}
	})

	t.Run("when command doesn't exist", func(t *testing.T) {
		gate := Gate{
			Name:    "does_not_exist",
			Command: "seofjseofjsoejfsoeif",
		}
		task := Task{
			TraceID: "ABC",
			Name:    "test",
		}

		gotB, gotR := gate.IsOpen(&task)

		if gotB != false {
			t.Errorf("got %v, want false", gotB)
		}

		if gotR.ExitCode != -1 {
			t.Errorf("exit code: got %d, want 1", gotR.ExitCode)
		}

		if string(gotR.StdoutBytes) != "" {
			t.Errorf("stdout: got %s, want %s", string(gotR.StdoutBytes), "")
		}

		if gotR.Duration != 0 {
			t.Errorf("duration: got %d, want 0", gotR.Duration)
		}

		if errors.Unwrap(gotR.Err) != exec.ErrNotFound {
			t.Errorf("error: got %s, want nil", errors.Unwrap(gotR.Err))
		}
	})
}
