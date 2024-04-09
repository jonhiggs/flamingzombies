package main

import (
	"fmt"
	"time"

	"git.altos/flamingzombies/db"
	fz "git.altos/flamingzombies/lib"
)

func main() {
	db.Start()
	config := fz.ReadConfig()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for _, t := range config.Tasks {
				if t.Ready(ts) {
					fmt.Println("running command ", t.Command, t.Hash())
					go t.Run()
				}
			}
		}
	}
}
