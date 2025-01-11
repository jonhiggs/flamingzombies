package config

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/command"
	"github.com/jonhiggs/flamingzombies/lib/log"
)

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
}

func (t *Task) Command() command.Cmd {
	return command.Cmd{
		Command: t.command,
		Args:    t.args,
		Envs:    t.envs,
		Dir:     Directory,
		Timeout: t.timeout(),
	}
}

func (t *Task) Description() string {
	return t.description
}

func (t *Task) ErrorNotifiers() []Notifier {
	return []Notifier{}
}

func (t *Task) Exec() bool {
	t.debug("execution began")
	result := t.Command().Exec()

	if result.Err != nil {
		// TODO(jh) 20250110: handle the error
		return false
	}

	return result.ExitCode == 0
}

func (t *Task) Frequency() time.Duration {
	return time.Duration(t.frequencySeconds) * time.Second
}

// Check that the task is in a valid state.
func (t *Task) IsValid() error {
	re := regexp.MustCompile(`^.+$`)
	if !re.Match([]byte(t.name)) {
		return fmt.Errorf("name '%s': %w", t.name, ErrInvalidName)
	}

	// The command string cannot be blank. When loading the configuration, a
	// better test is performed to make sure that the file actually exists.
	if len(t.command) < 1 {
		return fmt.Errorf("command '%s': %w", t.command, ErrCommandNotExist)
	}

	if t.frequencySeconds < 1 {
		return fmt.Errorf("freqency '%d': %w", t.frequencySeconds, ErrLessThan1)
	}

	if t.retryFrequencySeconds < 1 {
		return fmt.Errorf("retry_freqency '%d': %w", t.retryFrequencySeconds, ErrLessThan1)
	}

	if t.timeoutSeconds < 1 {
		return fmt.Errorf("timeout_seconds '%d': %w", t.timeoutSeconds, ErrLessThan1)
	}

	if t.timeoutSeconds > t.retryFrequencySeconds {
		return fmt.Errorf("timeout_seconds '%d': %w", t.retryFrequencySeconds, ErrTimeoutSlowerThanRetry)
	}

	if t.retries > 0 && t.retryFrequencySeconds > t.frequencySeconds {
		return fmt.Errorf("retry_requency '%d': %w", t.retryFrequencySeconds, ErrRetriesSlowerThanFrequency)
	}

	if t.priority < 1 {
		return fmt.Errorf("priority '%d': %w", t.priority, ErrLessThan1)
	}
	if t.priority > 99 {
		return fmt.Errorf("priority '%d': %w", t.priority, ErrGreaterThan99)
	}

	return nil
}

// step back though the data to find the previous state
func (t *Task) LastState() State {
	h := t.history >> t.retries
	m := t.historyMask >> t.retries

	mask := t.retryMask()

	for mask <= m {
		if h&mask == mask {
			return STATE_OK
		}

		if h&mask == 0 {
			return STATE_FAIL
		}

		h = h >> 1
		m = m >> 1
	}

	return STATE_UNKNOWN
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) Notifiers() []Notifier {
	return []Notifier{}
}

func (t *Task) Priority() int {
	return t.priority
}

func (t *Task) Ready(ts time.Time) bool {
	// the hash is used to spread the checks across time.
	// while the state is unknown, retry at the rate of RetryFrequencySeconds

	if t.State() == STATE_UNKNOWN {
		return (uint32(ts.Unix())+t.hash())%uint32(t.retryFrequencySeconds) == 0
	}

	return (uint32(ts.Unix())+t.hash())%uint32(t.frequencySeconds) == 0
}

func (t *Task) RecordStatus(b bool) {
	t.history = t.history << 1
	if b {
		t.history += 1
	}

	t.historyMask = t.historyMask << 1
	t.historyMask += 1

	switch t.State() {
	case STATE_OK:
		t.lastOk = time.Now()
	case STATE_FAIL:
		t.lastFail = time.Now()
	}
}

func (t *Task) Retries() int {
	return t.retries
}

func (t *Task) RetryFrequency() time.Duration {
	return time.Duration(t.retryFrequencySeconds) * time.Second
}

// Extract the current state from the history
func (t *Task) State() State {
	// if there aren't enough measurements, return STATE_UNKNOWN
	if t.retryMask() > t.historyMask {
		return STATE_UNKNOWN
	}

	v := t.history & t.retryMask()

	if v == 0 {
		return STATE_FAIL
	}

	if v == t.retryMask() {
		return STATE_OK
	}

	return STATE_UNKNOWN
}

// if the state changed
func (t *Task) StateChanged() bool {
	// if state is unknown, then we can't make an assessment.
	if t.State() == STATE_UNKNOWN {
		return false
	}

	// shift back to the last record. if we had the data to raise an alert,
	// then assume we did.
	l := (t.history >> 1) & t.retryMask()
	if l == (t.history & t.retryMask()) {
		return false
	}

	if t.LastState() == STATE_UNKNOWN {
		return false
	}

	return t.State() != t.LastState()
}

/// PRIVATE ///////////////////////////////////////////////////////////////////

// Create a checksum of a tasks configuration. The hash is used for a
// consistent execution offset. Offsetting the execution prevents the time that
// tasks are executed from clustering around each other.
func (t *Task) hash() uint32 {
	// To help with testing, return hash of zero when there isn't a command or
	// any arguments.
	if t.command == "" && len(t.args) == 0 {
		return uint32(0)
	}

	s := t.command
	for _, a := range t.args {
		s += a
	}

	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// Determine whether sufficifent retries have been performed?
func (t *Task) retryMask() uint32 {
	var m uint32
	for i := 0; i < t.retries; i++ {
		m = m << 1
		m += 1
	}

	return m
}

func (t *Task) timeout() time.Duration {
	return time.Duration(t.timeoutSeconds) * time.Second
}

// print a debug log with this tasks metadata included
func (t *Task) debug(msg string) {
	log.Debug(msg,
		"type", "task",
		"name", t.Name(),
	)
}
