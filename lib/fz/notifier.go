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

func (n Notifier) Execute(env []string) {
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

	if err != nil {
		Error(fmt.Errorf("notifier %s: %w", n.Name, err))
	}

	Logger.Debug("output",
		"notifier", n.Name,
		"stdout", strings.TrimSuffix(string(stdoutBytes), "\n"),
		"stderr", strings.TrimSuffix(string(stderrBytes), "\n"),
	)

	if err != nil {
		if os.IsPermission(err) {
			// XXX: This risks loops if an ErrorNotifier has invalid permissions.
			Error(fmt.Errorf("notifier %s: %w", n.Name, ErrInvalidPermissions))
		}

		if ctx.Err() == context.DeadlineExceeded {
			// XXX: This risks loops if an ErrorNotifier times out.
			Error(fmt.Errorf("notifier %s: %w", n.Name, ErrTimeout))
		}

		return
	}
}

func (n Notifier) Environment() []string {
	var v []string

	v = MergeEnvVars(v, n.Envs)
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}
