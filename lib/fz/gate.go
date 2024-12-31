package fz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"
)

// Return a bool describe the state of the gate. The task is required because
// in influences the environment used when invoking the gate's command.
func (g Gate) IsOpen(t *Task) (bool, CommandResult) {
	var r CommandResult
	r.TraceID = t.TraceID
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_GATE_TIMEOUT_SECONDS*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = cfg.Directory
	cmd.Env = g.environment(t)
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		r.Err = err
		r.ExitCode = -1
		return false, r
	}
	r.StdoutBytes, _ = io.ReadAll(stdout)
	r.StderrBytes, _ = io.ReadAll(stderr)
	err = cmd.Wait()
	r.Duration = time.Now().Sub(startTime)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
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
	var e []string

	e = MergeEnvVars(e, []string{
		fmt.Sprintf("GATE_NAME=%s", g.Name),
		fmt.Sprintf("GATE_TIMEOUT=%d", DEFAULT_GATE_TIMEOUT_SECONDS),
	})

	e = MergeEnvVars(e, g.Envs)

	// Fetch the task env vars if function was called with *Task as argument.
	t, ok := i.(*Task)
	if ok {
	}

	switch v := i.(type) {
	case *Task:
		e = MergeEnvVars(e, t.Environment())
	case Task:
		e = MergeEnvVars(e, t.Environment())
	case nil:
		// do nothing
	default:
		panic(fmt.Sprintf("cannot accept %v", v))
	}

	e = MergeEnvVars(e, cfg.Defaults.Envs)

	return e
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
