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
	TimeoutSeconds int `toml:"timeout_seconds"`
}

type Notification struct {
	Notifier *Notifier
	Task     *Task
	Version  uint64
}

var NotifyCh = make(chan Notification, 100)

func ProcessNotifications() {
	go func() {
		for {
		C:
			select {
			case n := <-NotifyCh:
				if n.Task.stateVersion > n.Version {
					log.WithFields(log.Fields{
						"file":                 "lib/notifier.go",
						"notifier_name":        n.Notifier.Name,
						"notification_version": n.Version,
						"task_state_version":   n.Task.stateVersion,
					}).Debug("skipping stale notification")
					break C
				}

				log.WithFields(log.Fields{
					"file":                 "lib/notifier.go",
					"notifier_name":        n.Notifier.Name,
					"notification_version": n.Version,
					"task_state_version":   n.Task.stateVersion,
				}).Info("sending notification")

				ctx, cancel := context.WithTimeout(context.Background(), n.Notifier.timeout())
				defer cancel()

				cmd := exec.CommandContext(ctx, n.Notifier.Command, n.Notifier.Args...)

				stdin, err := cmd.StdinPipe()
				if err != nil {
					log.WithFields(log.Fields{
						"file":                 "lib/notifier.go",
						"notifier_name":        n.Notifier.Name,
						"notification_version": n.Version,
						"task_state_version":   n.Task.stateVersion,
					}).Error(err)
				}
				defer stdin.Close()

				cmd.Env = []string{
					fmt.Sprintf("PRIORITY=%d", n.Task.Priority),
					fmt.Sprintf("SUBJECT=%s", n.subject()),
					fmt.Sprintf("STATE=%s", n.Task.State()),
					fmt.Sprintf("LAST_STATE=%s", n.Task.LastState()),
				}

				log.WithFields(log.Fields{
					"file":                 "lib/notifier.go",
					"notifier_name":        n.Notifier.Name,
					"notification_version": n.Version,
					"task_state_version":   n.Task.stateVersion,
				}).Trace(fmt.Sprintf("writing string to stdin: %s", n.body()))

				io.WriteString(stdin, n.body())

				err = cmd.Run()

				if ctx.Err() == context.DeadlineExceeded {
					log.WithFields(log.Fields{
						"file":                 "lib/notifier.go",
						"notifier_name":        n.Notifier.Name,
						"notification_version": n.Version,
						"task_state_version":   n.Task.stateVersion,
					}).Error(fmt.Sprintf("time out exceeded while executing notifier"))
				} else if err != nil {
					log.WithFields(log.Fields{
						"file":                 "lib/notifier.go",
						"notifier_name":        n.Notifier.Name,
						"notification_version": n.Version,
						"task_state_version":   n.Task.stateVersion,
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
		"task %s changed state from %s to %s",
		n.Task.Name,
		n.Task.LastState(),
		n.Task.State(),
	)
}

func (n Notification) body() string {
	if n.Task.State() == STATE_OK {
		return n.Task.RecoverBody
	} else if n.Task.State() == STATE_FAIL {
		return n.Task.ErrorBody
	}

	return fmt.Sprintf("The task %s is in an %s state", n.Task.Name, n.Task.State())
}
