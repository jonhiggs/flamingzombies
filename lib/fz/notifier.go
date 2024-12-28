package fz

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

func (n Notifier) Timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notifier) Execute(env []string) {
	ctx, cancel := context.WithTimeout(context.Background(), n.Timeout())
	defer cancel()

	cmd := exec.CommandContext(ctx, n.Command, n.Args...)
	cmd.Dir = cfg.Directory
	cmd.Env = env
	stderr, _ := cmd.StderrPipe()

	err := cmd.Run()

	if err != nil {
		if os.IsPermission(err) {
			Error(fmt.Errorf("notifier %s: %w", n.Name, ErrInvalidPermissions))
			return
		}

		if ctx.Err() == context.DeadlineExceeded {
			// XXX: This risks loops if an ErrorNotifier times out.
			Error(fmt.Errorf("notifier %s: %w", n.Name, ErrTimeout))
		}
	}

	errorMessage, _ := io.ReadAll(stderr)
	if len(errorMessage) > 0 {
		Logger.Debug(fmt.Sprintf("notifier returned stderr: %s", errorMessage), "notifier", n.Name)
	}
}

func (n Notifier) Environment() []string {
	var v []string

	for _, e := range cfg.Defaults.Envs {
		v = append(v, e)
	}

	for _, e := range n.Envs {
		v = append(v, e)
	}

	return v
}
