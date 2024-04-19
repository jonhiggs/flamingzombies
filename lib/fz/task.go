package fz

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Task struct {
	Name                  string   `toml:"name"`                    // friendly name
	Command               string   `toml:"command"`                 // command
	Args                  []string `toml:"args"`                    // command arguments
	FrequencySeconds      int      `toml:"frequency_seconds"`       // how often to run
	RetryFrequencySeconds int      `toml:"retry_frequency_seconds"` // how quickly to retry when state unknown
	TimeoutSeconds        int      `toml:"timeout_seconds"`         // how long an execution may run
	LockTimeoutSeconds    int      `toml:"lock_timeout_seconds"`    // how long to wait for a lock
	Retries               int      `toml:"retries"`                 // number of retries before changing the state
	NotifierNames         []string `toml:"notifiers"`               // notifiers to trigger upon state change
	Priority              int      `toml:"priority"`                // the priority of the notifications
	ErrorBody             string   `toml:"error_body"`              // the body of the notification when entering an error state
	RecoverBody           string   `toml:"recover_body"`            // the body of the notification when recovering from an error state

	history      uint32     // represented in binary. Successes are high
	measurements uint32     // the bits in the history with a recorded value. Needed to understand a history of 0
	mutex        sync.Mutex // lock to ensure one task runs at a time
}

func (t Task) Hash() uint32 {
	// To help with testing, return hash of zero when there isn't a command or
	// any arguments.
	if t.Command == "" && len(t.Args) == 0 {
		return uint32(0)
	}

	s := t.Command
	for _, a := range t.Args {
		s += a
	}

	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// how often to run
func (t Task) Frequency() time.Duration {
	if t.FrequencySeconds == 0 {
		return time.Duration(DEFAULT_FREQUENCY_SECONDS) * time.Second
	}

	return time.Duration(t.FrequencySeconds) * time.Second
}

func (t Task) Ready(ts time.Time) bool {
	// the hash is used to spread the checks across time.
	// if the state is unknown, retry at the rate of RetryFrequencySeconds

	if t.State() == STATE_UNKNOWN {
		return (uint32(ts.Unix())+t.Hash())%uint32(t.RetryFrequencySeconds) == 0
	}

	return (uint32(ts.Unix())+t.Hash())%uint32(t.FrequencySeconds) == 0
}

func (t *Task) Run() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("executing task")

	ctx, cancel := context.WithTimeout(context.Background(), t.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Error(fmt.Sprintf("time out exceeded while executing command"))

		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			log.WithFields(log.Fields{
				"file":      "lib/task.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Error(err)

			return false
		}

		exiterr, _ := err.(*exec.ExitError)
		panic(err)
		code := exiterr.ExitCode()

		//panic(fmt.Sprintf("status not 0: %d", code))
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Info(fmt.Sprintf("command exited with %d", code))
		t.RecordStatus(false)
	} else {
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Info(fmt.Sprintf("command succeeded"))

		t.RecordStatus(true)
	}

	if t.State() != STATE_UNKNOWN {
		for _, name := range t.NotifierNames {
			for i, n := range config.Notifiers {
				if n.Name == name {
					log.WithFields(log.Fields{
						"file":      "lib/task.go",
						"task_name": t.Name,
						"task_hash": t.Hash(),
					}).Debug(fmt.Sprintf("raising notification. is %s, was %s", t.State(), t.LastState()))
					NotifyCh <- Notification{&config.Notifiers[i], t}
				}
			}
		}
	}
	return true
}

func (t *Task) RecordStatus(b bool) {
	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Trace(fmt.Sprintf("recording status %v", b))

	t.history = t.history << 1
	if b {
		t.history += 1
	}

	t.measurements = t.measurements << 1
	t.measurements += 1

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Trace(fmt.Sprintf("history is %b", t.history))

}

// extract the current state from the history
func (t Task) State() State {
	// if there aren't enough measurements, return STATE_UNKNOWN
	if t.retryMask() > t.measurements {
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

// step back though the data to find the previous state
func (t Task) LastState() State {
	h := t.history >> t.Retries
	m := t.measurements >> t.Retries

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

// if the state changed
func (t Task) StateChanged() bool {
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

func (t Task) retryMask() uint32 {
	var m uint32
	for i := 0; i < t.Retries; i++ {
		m = m << 1
		m += 1
	}

	return m
}

func (t Task) timeout() time.Duration {
	return time.Duration(t.TimeoutSeconds) * time.Second
}

func (t Task) retryFrequency() time.Duration {
	return time.Duration(t.RetryFrequencySeconds) * time.Second
}

// TODO: this is currently unused
func (t Task) lockTimeout() time.Duration {
	return time.Duration(t.LockTimeoutSeconds) * time.Second
}

func (t Task) notifiers() []*Notifier {
	var not []*Notifier
	for _, nName := range t.NotifierNames {
		found := false
		for i, _ := range config.Notifiers {
			if nName == config.Notifiers[i].Name {
				not = append(not, &config.Notifiers[i])
				found = true
			}
		}

		if !found {
			log.WithFields(log.Fields{
				"file":      "lib/task.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal(fmt.Sprintf("unknown notifier '%s'", nName))
		}
	}

	return not
}

func (t Task) validate() error {
	if _, err := os.Stat(t.Command); os.IsNotExist(err) {
		return fmt.Errorf("task command not found")
	}

	return nil
}
