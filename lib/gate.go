package fz

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type Gate struct {
	Name           string   `toml:"name"`            // friendly name
	Command        string   `toml:"command"`         // command
	Args           []string `toml:"args"`            // command arguments
	TimeoutSeconds int      `toml:"timeout_seconds"` // how long an execution may run
}

func (g Gate) IsOpen(t *Task) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		log.WithFields(log.Fields{
			"file":      "lib/gate.go",
			"gate_name": g.Name,
		}).Error(fmt.Sprintf("time out exceeded while executing command"))

		return false
	}

	if err != nil {
		exiterr, _ := err.(*exec.ExitError)

		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"gate_name": g.Name,
		}).Debug(fmt.Sprintf("command exited with %d", exiterr.ExitCode()))

		return false
	}

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"gate_name": g.Name,
	}).Debug(fmt.Sprintf("command exited with %d", 0))
	return true
}
