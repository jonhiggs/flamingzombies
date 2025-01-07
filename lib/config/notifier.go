package config

import "time"

type Notifier struct {
	args           []string
	command        string
	envs           []string
	gateSetStrings [][]string
	name           string
	timeoutSeconds int
}

func (n Notifier) Args() []string {
	return n.args
}

func (n Notifier) Command() string {
	return n.command
}

func (n Notifier) Environment() []string {
	return n.envs
}

// the gates attached to the notifier
func (n Notifier) GateSets() [][]Gate {
	return [][]Gate{}
}

func (n Notifier) Name() string {
	return n.name
}

func (n Notifier) Timeout() time.Duration {
	return time.Duration(n.timeoutSeconds) * time.Second
}
