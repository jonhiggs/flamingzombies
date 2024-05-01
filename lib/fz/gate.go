package fz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type Gate struct {
	Name    string   `toml:"name"`    // friendly name
	Command string   `toml:"command"` // command
	Args    []string `toml:"args"`    // command arguments
}

func (g Gate) IsOpen(t *Task) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = config.Directory
	cmd.Env = []string{
		fmt.Sprintf("FREQUENCY=%d", t.FrequencySeconds),
		fmt.Sprintf("TASK_COMMAND=%s", t.Command),
		fmt.Sprintf("LAST_STATE=%s", t.LastState()),
		fmt.Sprintf("PRIORITY=%d", t.Priority),
		fmt.Sprintf("STATE=%s", t.State()),
		fmt.Sprintf("STATE_CHANGED=%v", t.StateChanged()),
		fmt.Sprintf("HISTORY=%d", t.History),
		fmt.Sprintf("HISTORY_MASK=%d", t.HistoryMask),
	}

	log.WithFields(log.Fields{
		"file":      "lib/fz/gate.go",
		"gate_name": g.Name,
	}).Trace(fmt.Sprintf("running"))

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		log.WithFields(log.Fields{
			"file":      "lib/fz/gate.go",
			"gate_name": g.Name,
		}).Error(fmt.Sprintf("time out exceeded while executing command"))

		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			log.WithFields(log.Fields{
				"file":      "lib/fz/gate.go",
				"gate_name": g.Name,
			}).Error(err)

			return false
		}

		exiterr, _ := err.(*exec.ExitError)

		log.WithFields(log.Fields{
			"file":      "lib/fz/gate.go",
			"gate_name": g.Name,
		}).Debug(fmt.Sprintf("command exited with %d", exiterr.ExitCode()))

		return false
	}

	log.WithFields(log.Fields{
		"file":      "lib/fz/gate.go",
		"gate_name": g.Name,
	}).Debug(fmt.Sprintf("command exited with %d", 0))
	return true
}
