package fz

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// Return a bool describe the state of the gate. The task is required because
// in influences the environment used when invoking the gate's command.
func (g Gate) IsOpen(t *Task) (bool, CommandResult) {
	var r CommandResult
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_GATE_TIMEOUT_SECONDS*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = cfg.Directory
	cmd.Env = g.environment(t)
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	r.StdoutBytes, _ = io.ReadAll(stdout)
	r.StderrBytes, _ = io.ReadAll(stderr)
	cmd.Wait()

	if err != nil {
		if os.IsPermission(err) {
			r.Err = ErrInvalidPermissions
		} else if ctx.Err() == context.DeadlineExceeded {
			r.Err = ErrTimeout
		} else {
			exiterr, _ := err.(*exec.ExitError)
			r.ExitCode = exiterr.ExitCode()
		}

		return false, r
	}

	return true, r
}

// return the environment needed when invoking a Gate for a Task.
func (g Gate) environment(i interface{}) []string {
	var v []string

	v = MergeEnvVars(v, []string{
		fmt.Sprintf("GATE_NAME=%s", g.Name),
		fmt.Sprintf("GATE_TIMEOUT=%d", DEFAULT_GATE_TIMEOUT_SECONDS),
	})

	v = MergeEnvVars(v, g.Envs)

	// Fetch the task env vars if function was called with *Task as argument.
	t, ok := i.(Task)
	if ok {
		v = MergeEnvVars(v, t.Environment())
	}

	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}

// check the state of a set of gates
func GateSetOpen(t *Task, gates ...*Gate) bool {
	for i, g := range gates {
		open, r := g.IsOpen(t)
		Logger.Debug("checking gate",
			"name", g.Name,
			"gateset_id", i,
			"open", fmt.Sprintf("%v", open),
			"stdout", string(r.StdoutBytes),
			"stderr", string(r.StderrBytes),
			"exit_code", r.ExitCode,
			"trace_id", t.TraceID,
		)

		// gateset is closed if any gate is closed
		if !open {
			return false
		}
	}
	return true
}
