package fz

import (
	"hash/fnv"
	"os/exec"
	"sync"
	"time"
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

func (t Task) Run() bool {
	if t.isLocked() {
		return true
	}

	cmd := exec.Command(t.Command, t.Args...)
	t.lock()

	err := cmd.Run()
	t.unlock()

	if _, ok := err.(*exec.ExitError); ok {
		StateRecordCh <- StateRecord{t.Hash(), false}
		return false
	}

	StateRecordCh <- StateRecord{t.Hash(), true}
	return true
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
	if !t.isLocked() {
		taskLocks = append(taskLocks, t.Hash())
	}

	return true
}

func (t Task) unlock() bool {
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
