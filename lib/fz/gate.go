package fz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (g Gate) IsOpen(t *Task, n *Notifier) bool {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_GATE_TIMEOUT_SECONDS*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = cfg.Directory
	cmd.Env = t.Environment()

	//startTime := time.Now()
	err := cmd.Run()
	//g.DurationMetric(time.Now().Sub(startTime))
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
