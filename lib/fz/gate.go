package fz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
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

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		Logger.Error(fmt.Sprintf("time out exceeded while executing gate"), "gate", g.Name)

		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			Logger.Error(fmt.Sprint(err), "gate", g.Name)
		}
		return false
	}
	return true
}

func GateByName(n string) (*Gate, error) {
	for i, g := range config.Gates {
		if g.Name == n {
			return &config.Gates[i], nil
		}
	}

	return nil, errors.New("named gate was not be found")
}

func (g Gate) validate() error {
	if _, err := os.Stat(g.Command); os.IsNotExist(err) {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, g.Command)); os.IsNotExist(err) {
			return fmt.Errorf("gate command not found")
		}
	}

	return nil
}
