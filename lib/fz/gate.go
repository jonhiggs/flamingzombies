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
		Logger.Error(fmt.Sprintf("time out exceeded while executing gate"), "gate", g.Name)
		//g.IncMetric("timeout")

		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			Logger.Error(fmt.Sprint(err), "gate", g.Name)
			//g.IncMetric("error")
		}

		//g.IncMetric("closed")
		return false
	}
	//g.IncMetric("open")
	return true
}
