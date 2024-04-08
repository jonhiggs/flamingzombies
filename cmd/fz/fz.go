package main

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/BurntSushi/toml"
)

type task struct {
	Command   string
	Args      []string
	Frequency int
	Timeout   int
	Running   bool
	State     int
}
type Tasks struct {
	Task []task
}

func main() {
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
	done := make(chan bool)
	go func() {
		time.Sleep(10 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			fmt.Println("Done!")
			return
		case ts := <-ticker.C:
			run(ts)
			for _, t := range tasks.Task {
				if t.Ready(ts) {
					go t.Run()
				}
			}
		}
	}
}

func run(t time.Time) {
	fmt.Println("Current time: ", t)
}

func (t task) Ready(ts time.Time) bool {
	return (uint32(ts.Second())+t.hash())%uint32(t.Frequency) == 0
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

func (t task) Run() bool {
	if t.Running {
		fmt.Println("task %s is still running: ", t.Command)
		return false
	}

	fmt.Println("Running command: %s", t.Command)
	t.Running = true

	cmd := exec.Command(t.Command, t.Args...)
	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		exiterr, _ := err.(*exec.ExitError)
		t.State = exiterr.ExitCode()
	}

	t.Running = false
	return true
}
