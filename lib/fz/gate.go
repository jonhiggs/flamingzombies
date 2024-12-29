package fz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (g Gate) Execute(t *Task) bool {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_GATE_TIMEOUT_SECONDS*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = cfg.Directory
	cmd.Env = append(g.Environment(), t.Environment()...)

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		Error(fmt.Errorf("gate %s: %w", g.Name, ErrTimeout), true)
		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			Error(fmt.Errorf("task %s: %w", g.Name, ErrInvalidPermissions), true)
		}
		return false
	}

	return true
}

func (g Gate) Environment() []string {
	var v []string

	v = MergeEnvVars(v, g.Envs)
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}
