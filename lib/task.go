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

var unlockLock sync.Mutex // ensure two unlocks don't run concurrently

type Task struct {
	Name               string   `toml:"name"`                 // friendly name
	Command            string   `toml:"command"`              // command
	Args               []string `toml:"args"`                 // command aruments
	FrequencySeconds   int      `toml:"frequency_seconds"`    // how often to run
	TimeoutSeconds     int      `toml:"timeout_seconds"`      // how long an execution may run
	LockTimeoutSeconds int      `toml:"lock_timeout_seconds"` // how long to wait for a lock
	Retries            int      `toml:"retries"`              // historic values used to determine the status

	LockTimout  time.Duration // how long to wait for a lock
	Timeout     time.Duration // how long an execution may run (for system)
	NotifierStr []string      // notifiers to trigger upon state change
	Notifiers   []*Notifier   // notifiers to trigger upon state change
	history     uint32        // represented in binary. sucessess are high
	mutex       sync.Mutex    // lock to ensure one task runs at a time
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
	return (uint32(ts.Unix())+t.Hash())%uint32(t.FrequencySeconds) == 0
}

func (t *Task) Run() bool {
	// TODO: add a deadline
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("executing task")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
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

	return false

	originalState := t.State()
	t.RecordStatus(true)
	newState := t.State()

	if newState != originalState {
		for _, n := range t.Notifiers {
			NotifyCh <- Notification{
				Notifier: n,
				Subject:  fmt.Sprintf("command %s changed from %d to %d", t.Name, originalState, newState),
				Body:     "blah, blah, blah",
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
