package fz

import (
	"fmt"
	"hash/fnv"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var unlockLock sync.Mutex // ensure two unlocks don't run concurrently

type Task struct {
	Name       string   // friendly name
	Command    string   // command
	Args       []string // command aruments
	Frequency  int      // how often to run
	Timeout    int      // how long an execution may run
	LockTimout int      // how long to wait for a lock
	Notifiers  []string
	Retries    int
	History    uint32     // represented in binary. sucessess are high
	Mutex      sync.Mutex // lock to ensure one task runs at a time
}

func (t Task) Hash() uint32 {
	s := t.Command
	for _, a := range t.Args {
		s += a
	}

	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func (t Task) Ready(ts time.Time) bool {
	return (uint32(ts.Second())+t.Hash())%uint32(t.Frequency) == 0
}

func (t *Task) Run() bool {
	// TODO: add a deadline
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("executing task")
	cmd := exec.Command(t.Command, t.Args...)

	err := cmd.Run()

	if _, ok := err.(*exec.ExitError); ok {
		t.RecordStatus(false)
		return false
	}

	t.RecordStatus(true)
	return true
}

func (t *Task) RecordStatus(b bool) {
	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("recording status")

	t.History = t.History << 1
	if b {
		t.History += 1
	}

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Trace(fmt.Sprintf("history is %b", t.History))
}

// extract the current state from the History
func (t *Task) State() int {
	return -1
}
