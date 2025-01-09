package core

import "github.com/jonhiggs/flamingzombies/lib/config"

// Process the scheduled tasks

var taskQueue = make(chan *config.Task, 100)

func ProcessTasks() {
	for {
		select {
		case t := <-taskQueue:
			executeTask(t)
			// TODO(jh) 20250110: the work
		}
	}
}

/// PRIVATE ///////////////////////////////////////////////////////////////////

func executeTask(t *config.Task) {
	// TODO(jh) 20250110: add locking
	// TODO(jh) 20250110: record the result
	// TODO(jh) 20250110: update the last* timestamps
	// TODO(jh) 20250110: trigger the notifier
}
