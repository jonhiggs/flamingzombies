package fz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const DEFAULT_MIN_PRIORITY = 100

type Notifier struct {
	Name           string
	Command        string
	Args           []string
	MinPriority    int
	TimeoutSeconds int `toml:"timeout_seconds"`
}

type Notification struct {
	Notifier *Notifier
	Task     *Task
}

var NotifyCh = make(chan Notification, 100)

func ProcessNotifications() {
	go func() {
		for {
		C:
			select {
			case n := <-NotifyCh:
				if n.Notifier.MinPriority != 0 && n.Task.Priority > n.Notifier.MinPriority { // 1 is a higher priority than 2
					log.WithFields(log.Fields{
						"file":          "lib/notifier.go",
						"notifier_name": n.Notifier.Name,
					}).Info(fmt.Sprintf("not notifying because notification priority (%d) is lower than the notifiers minimum_priority (%d)", n.Task.Priority, n.Notifier.MinPriority))
					break C
				}

				log.WithFields(log.Fields{
					"file":          "lib/notifier.go",
					"notifier_name": n.Notifier.Name,
				}).Info("sending notification")

				ctx, cancel := context.WithTimeout(context.Background(), n.Notifier.timeout())
				defer cancel()

				cmd := exec.CommandContext(ctx, n.Notifier.Command, n.Notifier.Args...)

				stdin, err := cmd.StdinPipe()
				if err != nil {
					log.WithFields(log.Fields{
						"file":          "lib/notifier.go",
						"notifier_name": n.Notifier.Name,
					}).Error(err)
				}
				defer stdin.Close()

				cmd.Env = []string{
					fmt.Sprintf("PRIORITY=%d", n.Task.Priority),
					fmt.Sprintf("SUBJECT=%s", n.subject()),
					fmt.Sprintf("STATE=%d", n.Task.State()),
				}

				log.WithFields(log.Fields{
					"file":          "lib/notifier.go",
					"notifier_name": n.Notifier.Name,
				}).Trace(fmt.Sprintf("writing string to stdin: %s", n.body()))

				io.WriteString(stdin, n.body())

				if err := cmd.Run(); err != nil {
					log.WithFields(log.Fields{
						"file":          "lib/notifier.go",
						"notifier_name": n.Notifier.Name,
					}).Error(err)
				}

			}
		}
	}()
}

func (n Notifier) timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notification) subject() string {
	return fmt.Sprintf(
		"task %s changed state from %d to %d",
		n.Task.Name,
		n.Task.lastState,
		n.Task.State(),
	)
}

func (n Notification) body() string {
	if n.Task.State() == STATE_OK {
		return n.Task.RecoverBody
	} else if n.Task.State() == STATE_FAIL {
		return n.Task.ErrorBody
	}

	return fmt.Sprintf("The task %s is in an unknown state", n.Task.Name)
}
