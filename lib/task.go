package fz

import (
	"context"
	"fmt"
	"hash/fnv"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const STATE_UNKNOWN = -1
const STATE_FAIL = 0
const STATE_OK = 1

var unlockLock sync.Mutex // ensure two unlocks don't run concurrently

type Task struct {
	Name                  string   `toml:"name"`                    // friendly name
	Command               string   `toml:"command"`                 // command
	Args                  []string `toml:"args"`                    // command aruments
	FrequencySeconds      int      `toml:"frequency_seconds"`       // how often to run
	RetryFrequencySeconds int      `toml:"retry_frequency_seconds"` // how quickly to retry when state unknown
	TimeoutSeconds        int      `toml:"timeout_seconds"`         // how long an execution may run
	LockTimeoutSeconds    int      `toml:"lock_timeout_seconds"`    // how long to wait for a lock
	Retries               int      `toml:"retries"`                 // number of retries before changing the state
	NotifierNames         []string `toml:"notifiers"`               // notifiers to trigger upon state change
	Priority              int      `toml:"priority"`                // the priority of the notifications
	ErrorBody             string   `toml:"error_body"`              // the body of the notifiation when entering an error state
	RecoverBody           string   `toml:"recover_body"`            // the body of the notifiation when recovering from an error state

	history      uint32     // represented in binary. sucessess are high
	lastState    int        // last known state
	stateChanged bool       // set to true if the last measure changed the state
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

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Error("command failed or timed out")

		t.RecordStatus(false)
		return false
	}

	t.RecordStatus(true)
	if t.stateChanged {
		for _, n := range t.notifications() {
			NotifyCh <- n
		}
	}
	return true
}

func (t *Task) RecordStatus(b bool) {
	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("recording status")

	t.history = t.history << 1
	if b {
		t.history += 1
	}

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Trace(fmt.Sprintf("history is %b", t.history))

	s := t.State()
	if (s != STATE_UNKNOWN) && (s != t.lastState) {
		t.lastState = s
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Trace(fmt.Sprintf("updating lastState to %d", s))
		t.stateChanged = true
	} else {
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Trace("state was unchanged")
		t.stateChanged = false
	}
}

// extract the current state from the history
// the returned values are:
//
//	-1: unknown
//	 0: down
//	 1: up
func (t *Task) State() int {
	var mask uint32
	for i := 0; i < t.Retries; i++ {
		mask = mask << 1
		mask += 1
	}

	v := t.history & mask

	if v == 0 {
		return 0
	}

	if v == mask {
		return 1
	}

	return -1
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

func (t Task) notifications() []Notification {
	var ns []Notification
	var body string

	if t.State() == STATE_OK {
		body = t.ErrorBody
	} else {
		body = t.RecoverBody
	}

	subject := fmt.Sprintf("task %s changed state from %d to %d", t.Name, t.lastState, t.State())

	for _, n := range t.notifiers() {
		ns = append(ns, Notification{
			Notifier: n,
			Subject:  subject,
			Body:     body,
			Priority: t.Priority,
		})
	}

	return ns
}
