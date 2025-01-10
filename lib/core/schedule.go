package core

import (
	"time"
)

// queue work to be processed

// Every second, check for tasks that are ready ready for scheduling.
func ScheduleTasks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for i, t := range cfg.Tasks {
				if t.Ready(ts) {
					taskQueue <- t
				}
			}
		}
	}
}
