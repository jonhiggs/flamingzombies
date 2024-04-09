package fz

import (
	"fmt"
	"hash/fnv"
	"os/exec"
	"time"

	"git.altos/flamingzombies/db"
)

type Task struct {
	Name      string
	Command   string
	Args      []string
	Frequency int
	Timeout   int
	State     int
	Notifier  string
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
		fmt.Println("waiting for lock to release")
		return true
	}

	fmt.Printf("Running command: %s\n", t.Command)

	t.lock()
	cmd := exec.Command(t.Command, t.Args...)
	cmd.Run()
	t.unlock()

	return true
}

func (t Task) isLocked() bool {
	return db.IsLocked(t.Hash())
}

func (t Task) lock() bool {
	db.LockCh <- db.Lock{t.Hash(), true}
	return true
}

func (t Task) unlock() bool {
	db.LockCh <- db.Lock{t.Hash(), false}
	return true
}
