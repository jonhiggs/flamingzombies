package fz

import (
	"fmt"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/run"
)

// Return a bool describe the state of the gate. The task is required because
// in influences the environment used when invoking the gate's command.
func (g Gate) IsOpen(t *Task) (bool, run.Result) {
	c := run.Cmd{
		Command: g.Command,
		Args:    g.Args,
		Envs:    g.environment(t),
		Dir:     cfg.Directory,
		TraceID: t.TraceID,
		Timeout: DEFAULT_GATE_TIMEOUT_SECONDS * time.Second,
	}

	r := c.Start()

	return r.ExitCode == 0, r
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
