package config

import (
	"time"

	"github.com/jonhiggs/flamingzombies/lib/command"
)

type Notifier struct {
	args           []string
	command        string
	envs           []string
	gateSetStrings [][]string
	name           string
	timeoutSeconds int
}

func (n Notifier) Command() command.Cmd {
	return command.Cmd{
		Command: n.command,
		Args:    n.args,
		Envs:    n.envs,
		Dir:     Directory,
		Timeout: time.Duration(n.timeoutSeconds) * time.Second,
	}
}

func (n Notifier) EvaluateGates() bool {
	for _, gs := range n.GateSets() {
		for _, g := range gs {
			if !g.Exec() {
				return false
			}
		}

		// this gateset is open
		return true
	}

	// no more gatesets to check
	return false
}

func (n *Notifier) Exec() bool {
	result := n.Command().Exec()

	if result.Err != nil {
		// TODO(jh) 20250110: handle the error
		return false
	}

	return result.ExitCode == 0
}

// the gates attached to the notifier
func (n Notifier) GateSets() [][]Gate {
	return [][]Gate{}
}

func (n Notifier) Name() string {
	return n.name
}
