package core

import (
	"time"

	"github.com/jonhiggs/flamingzombies/lib/config"
)

// queue work to be processed

// Every second, check for tasks that are ready ready for scheduling.
func ScheduleTasks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for _, t := range config.Tasks {
				if t.Ready(ts) {
					taskQueue <- &t
				}
			}
		}
	}
}
