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

func (g Gate) Execute(t *Task) bool {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_GATE_TIMEOUT_SECONDS*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = cfg.Directory
	cmd.Env = append(g.Environment(), t.Environment()...)
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	stdoutBytes, _ := io.ReadAll(stdout)
	stderrBytes, _ := io.ReadAll(stderr)
	cmd.Wait()

	Logger.Debug("output",
		"gate", g.Name,
		"task", t.Name,
		"stdout", strings.TrimSuffix(string(stdoutBytes), "\n"),
		"stderr", strings.TrimSuffix(string(stderrBytes), "\n"),
	)

	if err != nil {
		if os.IsPermission(err) {
			Error(fmt.Errorf("gate %s: %w", g.Name, ErrInvalidPermissions), true)
		} else if ctx.Err() == context.DeadlineExceeded {
			Error(fmt.Errorf("gate %s: %w", g.Name, ErrTimeout), true)
		} else {
			Error(fmt.Errorf("gate %s: %w", g.Name, err), true)
		}

		return false
	}

	return true
}

func (g Gate) Environment(tasks ...Task) []string {
	var v []string

	v = MergeEnvVars(v, []string{
		fmt.Sprintf("GATE_NAME=%s", g.Name),
		fmt.Sprintf("GATE_TIMEOUT=%d", DEFAULT_GATE_TIMEOUT_SECONDS),
	})

	v = MergeEnvVars(v, g.Envs)
	for _, t := range tasks {
		v = MergeEnvVars(v, t.Environment())
	}
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}
