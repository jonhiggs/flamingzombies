package fz

import (
	"fmt"
	"hash/fnv"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// list of task hashes that are locked
var taskLocks []uint32
var unlockLock sync.Mutex // ensure two unlocks don't run concurrently

type Task struct {
	Name      string
	Command   string
	Args      []string
	Frequency int
	Timeout   int
	State     int
	Notifiers []string
	Retries   int
	History   uint32 // represented in binary. sucessess are high
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
	if t.isLocked() {
		log.WithFields(log.Fields{
			"file":      "lib/task.go",
			"task_name": t.Name,
			"task_hash": t.Hash(),
		}).Info("aborting task because it is locked")
		return true
	}

	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("executing task")
	cmd := exec.Command(t.Command, t.Args...)
	t.lock()
	defer t.unlock()

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

func (t Task) isLocked() bool {
	for _, h := range taskLocks {
		if h == t.Hash() {
			return true
		}
	}

	return false
}

func (t Task) lock() bool {
	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("locking")
	if !t.isLocked() {
		taskLocks = append(taskLocks, t.Hash())
	}

	return true
}

func (t Task) unlock() bool {
	log.WithFields(log.Fields{
		"file":      "lib/task.go",
		"task_name": t.Name,
		"task_hash": t.Hash(),
	}).Info("unlocking")
	unlockLock.Lock()
	defer unlockLock.Unlock()
	newLocks := []uint32{}
	for _, h := range taskLocks {
		if h != t.Hash() {
			newLocks = append(newLocks, h)
		}
	}
	taskLocks = newLocks
	return true
}
