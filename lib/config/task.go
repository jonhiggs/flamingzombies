package config

import "time"

// A Task is a command that is executed on a schedule. The struct contains the
// static configuration of the task which is read from the configuration file,
// and it's metadata and history which are generated over the course of the
// daemons lifecycle.
type Task struct {
	args                  []string   // command arguments
	command               string     // command
	description           string     // description of the task
	envs                  []string   // environment variables supplied to task
	errorNotifierNames    []string   // notifiers to trigger upon state change
	frequencySeconds      int        // how often to run
	history               uint32     // represented in binary. Successes are high
	historyMask           uint32     // the bits in the history with a recorded value. Needed to understand a history of 0
	lastFail              time.Time  // the time of the last failed execution
	lastNotification      time.Time  // the time of the last notification
	lastOk                time.Time  // the time of the last successful execution
	lastRun               time.Time  // the time of the last execution
	name                  string     // friendly name
	notifierNames         []string   // notifiers to trigger upon state change
	notifiers             []Notifier // the child notifiers of this task
	priority              int        // the priority of the notifications
	retries               int        // number of retries before changing the state
	retryFrequencySeconds int        // how quickly to retry when state unknown
	timeoutSeconds        int        // how long an execution may run
	traceID               string     // the ID of the task execution to help with tracing
}

func (t Task) Args() []string {
	return t.args
}

func (t Task) Command() string {
	return t.command
}

func (t Task) Description() string {
	return t.description
}

func (t Task) Environment() []string {
	return t.envs
}

func (t Task) ErrorNotifiers() []Notifier {
	return []Notifier{}
}

func (t Task) Frequency() time.Duration {
	return time.Duration(t.frequencySeconds) * time.Second
}

func (t Task) Name() string {
	return t.name
}

func (t Task) Notifiers() []Notifier {
	return []Notifier{}
}

func (t Task) Priority() int {
	return t.priority
}

func (t Task) Retries() int {
	return t.retries
}

func (t Task) RetryFrequency() time.Duration {
	return time.Duration(t.retryFrequencySeconds) * time.Second
}

func (t Task) Timeout() time.Duration {
	return time.Duration(t.timeoutSeconds) * time.Second
}
