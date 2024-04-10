package main

import (
	"time"

	"git.altos/flamingzombies/db"
	fz "git.altos/flamingzombies/lib"
)

func main() {
	config := fz.ReadConfig()

	db.Start()
	fz.ProcessNotifications()
	fz.RecordStates()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for _, t := range config.Tasks {
				if t.Ready(ts) {
					go t.Run()
				}
			}
		}
	}
}
