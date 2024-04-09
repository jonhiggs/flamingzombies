package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"git.altos/flamingzombies/db"
	fz "git.altos/flamingzombies/lib"

	"github.com/BurntSushi/toml"
)

func main() {
	db.Start()

	taskToml, err := ioutil.ReadFile("tasks.toml") // the file is inside the local directory
	if err != nil {
		fmt.Println("Err: %s", err)
	}

	var tasks fz.Tasks
	if _, err := toml.Decode(string(taskToml), &tasks); err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for _, t := range tasks.Task {
				if t.Ready(ts) {
					fmt.Println("running command ", t.Command, t.Hash())
					go t.Run()
				}
			}
		}
	}
}
