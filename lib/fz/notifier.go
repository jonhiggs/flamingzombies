package fz

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func (n Notifier) Timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notifier) Execute(traceID string, env []string, notifyErrors bool) {
	ctx, cancel := context.WithTimeout(context.Background(), n.Timeout())
	defer cancel()

	cmd := exec.CommandContext(ctx, n.Command, n.Args...)
	cmd.Dir = cfg.Directory
	cmd.Env = env
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)
	cmd.Wait()

	Logger.Debug("output",
		"notifier", n.Name,
		"stdout", strings.TrimSuffix(string(stdoutBytes), "\n"),
		"stderr", strings.TrimSuffix(string(stderrBytes), "\n"),
		"trace_id", traceID,
	)

	if err != nil {
		if os.IsPermission(err) {
			Error(traceID, fmt.Errorf("notifier %s: %w", n.Name, ErrInvalidPermissions), notifyErrors)
		} else if ctx.Err() == context.DeadlineExceeded {
			Error(traceID, fmt.Errorf("notifier %s: %w", n.Name, ErrTimeout), notifyErrors)
		} else {
			Error(traceID, fmt.Errorf("notifier %s: %w", n.Name, err), notifyErrors)
		}
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
