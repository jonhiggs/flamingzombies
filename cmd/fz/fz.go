package main

import (
	"time"

	"git.altos/flamingzombies/db"
	fz "git.altos/flamingzombies/lib"
)

var config fz.Config

func init() {
	config = fz.ReadConfig()
	fz.StartLogger(config.LogLevel)
	db.Start()
	fz.ProcessNotifications()
	fz.RecordStates()
}

func main() {
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
