package core

import (
	"fmt"

	"github.com/jonhiggs/flamingzombies/lib/config"
	"github.com/jonhiggs/flamingzombies/lib/trace"
)

// Process the scheduled tasks

var taskQueue = make(chan *config.Task, 100)

func ProcessTasks() {
	for {
		select {
		case t := <-taskQueue:
			executeTask(t)
		}
	}
}

/// PRIVATE ///////////////////////////////////////////////////////////////////

func executeTask(t *config.Task) {
	// TODO(jh) 20250110: add locking

	id = trace.New()

	ok := t.Exec(id)
	t.RecordStatus(ok)

	for _, n := range t.Notifiers() {
		if n.IsUngated {
			n.SetDescription("the description")
			n.SetMessage("the message")
			n.SetPriority(1)
			n.SetSubject("the subject")
			n.SetTraceID(id)

			_, err := n.Exec()
			if err != nil {
				t.LogError(fmt.Sprintf("%s", err), id)
			}
		}
	}
}
