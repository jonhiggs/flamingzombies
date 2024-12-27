package fz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (g Gate) Execute(env []string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_GATE_TIMEOUT_SECONDS*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = cfg.Directory
	cmd.Env = env

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		Error(fmt.Errorf("gate %s: %w", g.Name, ErrTimeout))
		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			Error(fmt.Errorf("task %s: %w", g.Name, ErrInvalidPermissions))
		}
		return false
	}

	return true
}
