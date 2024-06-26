package fz

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Gate struct {
	Name    string   `toml:"name"`    // friendly name
	Command string   `toml:"command"` // command
	Args    []string `toml:"args"`    // command arguments
}

func (g Gate) IsOpen(t *Task, n *Notifier) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, g.Command, g.Args...)

	cmd.Dir = config.Directory
	cmd.Env = []string{
		fmt.Sprintf("FREQUENCY=%d", t.FrequencySeconds),
		fmt.Sprintf("HISTORY=%d", t.History),
		fmt.Sprintf("HISTORY_MASK=%d", t.HistoryMask),
		fmt.Sprintf("LAST_FAIL=%d", t.LastFail.Unix()),
		fmt.Sprintf("LAST_NOTIFICATION=%d", t.GetLastNotification(n.Name).Unix()),
		fmt.Sprintf("LAST_OK=%d", t.LastOk.Unix()),
		fmt.Sprintf("LAST_STATE=%s", t.LastState()),
		fmt.Sprintf("PRIORITY=%d", t.Priority),
		fmt.Sprintf("STATE=%s", t.State()),
		fmt.Sprintf("STATE_CHANGED=%v", t.StateChanged()),
		fmt.Sprintf("TASK_COMMAND=%s", t.Command),
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

func (g Gate) validate() error {
	if _, err := os.Stat(g.Command); os.IsNotExist(err) {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, g.Command)); os.IsNotExist(err) {
			return fmt.Errorf("gate command not found")
		}
	}

	if strings.ContainsRune(g.Name, ' ') {
		return fmt.Errorf("name cannot contain spaces")
	}

	return nil
}

func GateByName(name string) (*Gate, error) {
	for i, g := range config.Gates {
		if g.Name == name {
			return &config.Gates[i], nil
		}
	}

	return nil, fmt.Errorf("gate '%s' is not known", name)
}
