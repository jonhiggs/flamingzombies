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

func (g Gate) Command(t string) command.Cmd {
	return command.Cmd{
		Command: g.command,
		Args:    g.args,
		Envs:    g.envs,
		Dir:     Directory,
		TraceID: t,
		Timeout: GATE_TIMEOUT,
	}
}
