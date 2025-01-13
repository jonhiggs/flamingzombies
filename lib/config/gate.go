package config

import (
	"time"

	"github.com/jonhiggs/flamingzombies/lib/command"
)

const GATE_TIMEOUT = 1 * time.Second

type Gate struct {
	args    []string // command arguments
	command string   // command
	envs    []string // environment variables
	name    string   // friendly name
}

func (g Gate) Name() string {
	return g.name
}

func (g Gate) Command() command.Cmd {
	return command.Cmd{
		Command: g.command,
		Args:    g.args,
		Envs:    g.envs,
		Dir:     Directory,
		Timeout: GATE_TIMEOUT,
	}
}

// return true of the gate is open
func (g *Gate) Exec() bool {
	result := g.Command().Exec()

	if result.Err != nil {
		// TODO(jh) 20250110: handle the error
		return false
	}

	return result.ExitCode == 0
}

func (g *Gate) IsValid() error {
	return nil
}
