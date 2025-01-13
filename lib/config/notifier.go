package config

import (
	"fmt"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/command"
	"github.com/jonhiggs/flamingzombies/lib/log"
	"github.com/jonhiggs/flamingzombies/lib/trace"
)

type Notifier struct {
	args           []string
	command        string
	envs           []string
	gateSetStrings [][]string
	name           string
	timeoutSeconds int
}

func (n *Notifier) Command() command.Cmd {
	return command.Cmd{
		Command: n.command,
		Args:    n.args,
		Envs:    n.envs,
		Dir:     Directory,
		Timeout: time.Duration(n.timeoutSeconds) * time.Second,
	}
}

func (n *Notifier) IsUngated() bool {
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

func (n *Notifier) IsValid() error {
	return nil
}

func (n *Notifier) Exec() (command.Result, error) {
	result := n.Command().Exec()

	if result.Err != nil {
		return result, result.Err
	}

	if result.ExitCode != 0 {
		return result, ErrNonZero
	}

	return result, nil
}

func (n *Notifier) SetDescription(s string) {
	n.envs = mergeEnvVars([]string{fmt.Sprintf("DESCRIPTION=%s", s)}, n.envs)
}

func (n *Notifier) SetMessage(s string) {
	n.envs = mergeEnvVars([]string{fmt.Sprintf("MESSAGE=%s", s)}, n.envs)
}

func (n *Notifier) SetPriority(i int) {
	n.envs = mergeEnvVars([]string{fmt.Sprintf("PRIORITY=%d", i)}, n.envs)
}

func (n *Notifier) SetSubject(s string) {
	n.envs = mergeEnvVars([]string{fmt.Sprintf("SUBJECT=%s", s)}, n.envs)
}

func (n *Notifier) SetTraceID(id trace.ID) {
	n.envs = mergeEnvVars([]string{fmt.Sprintf("TRACE_ID=%s", id)}, n.envs)
}

// the gates attached to the notifier
func (n Notifier) GateSets() [][]Gate {
	return [][]Gate{}
}

func (n Notifier) Name() string {
	return n.name
}

/// LOGGING ///////////////////////////////////////////////////////////////////

// Print a debug log.
func (n *Notifier) LogDebug(msg string, id trace.ID) {
	log.Debug(msg, n.LogData(id)...)
}

// The metadata to attach to all the logs
func (n *Notifier) LogData(id trace.ID) []any {
	return []any{
		"name", n.Name(),
		"trace_id", id,
		"type", "task",
	}
}
