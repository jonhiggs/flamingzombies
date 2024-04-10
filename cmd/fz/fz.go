package main

import (
	"time"

	fz "git.altos/flamingzombies/lib"
)

var config fz.Config

func init() {
	config = fz.ReadConfig()
	fz.StartLogger(config.LogLevel)
	fz.ProcessNotifications()
	fz.RecordStates()
}

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for i, t := range config.Tasks {
				if t.Ready(ts) {
					go config.Tasks[i].Run()
				}
			}
		}
	}
}
