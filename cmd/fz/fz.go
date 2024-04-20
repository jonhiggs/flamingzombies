package main

import (
	"time"

	"git.altos/flamingzombies/lib/daemon"
	"git.altos/flamingzombies/lib/fz"
)

var config fz.Config

func init() {
	config = fz.ReadConfig()
	fz.StartLogger(config.LogLevel)
	fz.ProcessNotifications()
	go daemon.Listen(&config)

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
