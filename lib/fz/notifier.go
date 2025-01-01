package fz

import (
	"fmt"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/run"
)

func (n Notifier) Timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notifier) Execute(traceID string, env []string, notifyErrors bool) {
	c := run.Cmd{
		Command: n.Command,
		Args:    n.Args,
		Envs:    env,
		Dir:     cfg.Directory,
		TraceID: traceID,
		Timeout: n.Timeout(),
	}

	r := c.Start()

	Logger.Debug("cmd result",
		"notifier", n.Name,
		"code", r.ExitCode,
		"stdout", r.Stdout(),
		"stderr", r.Stderr(),
		"trace_id", traceID,
	)

	if r.ExitCode < 0 {
		Error(traceID, fmt.Errorf("notifier %s: %w", n.Name, r.Err), notifyErrors)
	}
}

// Resolve the *Gates from the GateSetStrings
func (n Notifier) GateSets() [][]*Gate {
	r := [][]*Gate{}

	for _, gateSet := range n.GateSetStrings {
		gs := []*Gate{}
		for _, gateName := range gateSet {
			gs = append(gs, cfg.GetGateByName(gateName))
		}
		r = append(r, gs)
	}

	return r
}

func (n Notifier) Environment() []string {
	var v []string

	v = MergeEnvVars(v, n.Envs)
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}
