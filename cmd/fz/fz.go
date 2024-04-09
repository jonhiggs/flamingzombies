package main

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"git.altos/flamingzombies/db"

	"github.com/BurntSushi/toml"
)

type task struct {
	Command   string
	Args      []string
	Frequency int
	Timeout   int
	State     int
}
type Tasks struct {
	Task []task
}

func main() {
	db.Start()

	taskToml, err := ioutil.ReadFile("tasks.toml") // the file is inside the local directory
	if err != nil {
		fmt.Println("Err: %s", err)
	}

	var tasks Tasks
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
					fmt.Println("running command ", t.Command, t.hash())
					go t.Run()
				}
			}
		}
	}
}

func (t task) Ready(ts time.Time) bool {
	return (uint32(ts.Second())+t.hash())%uint32(t.Frequency) == 0
}

func (t task) Run() bool {
	if db.IsLocked(t.hash()) {
		fmt.Println("waiting for lock to release")
		return true
	}

	fmt.Printf("Running command: %s\n", t.Command)

	db.LockCh <- db.Lock{t.hash(), true}
	cmd := exec.Command(t.Command, t.Args...)
	//err := cmd.Run()
	cmd.Run()
	db.LockCh <- db.Lock{t.hash(), false}
	//exiterr, _ := err.(*exec.ExitError)
	//t.State = exiterr.ExitCode()

	return true
}

func (t task) hash() uint32 {
	s := t.Command
	for _, a := range t.Args {
		s += a
	}

	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
