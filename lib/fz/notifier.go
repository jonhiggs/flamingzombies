package fz

import (
	"context"
	"fmt"
	"io"
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
		if ctx.Err() == context.DeadlineExceeded {
			Logger.Error(fmt.Sprintf("timeout exceeded while executing notifier"), "notifier", n.Name)
			for _, errN := range cfg.Defaults.ErrorNotifierNames {
				ErrorNotifyCh <- ErrorNotification{
					Notifier: cfg.GetNotifierByName(errN),
					Error:    err,
				}
			}
		} else {
			// TODO
		}
	}

	errorMessage, _ := io.ReadAll(stderr)
	if len(errorMessage) > 0 {
		Logger.Debug(fmt.Sprintf("notifier returned stderr: %s", errorMessage), "notifier", n.Name)
	}
}
