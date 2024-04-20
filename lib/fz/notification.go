package fz

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

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
				for _, g := range n.Notifier.gates() {
					if g.IsOpen(n.Task) == false {
						log.WithFields(log.Fields{
							"file":          "lib/fz/notification.go",
							"notifier_name": n.Notifier.Name,
							"gate_name":     g.Name,
						}).Debug(fmt.Sprintf("gate is closed"))

						break C
					}
				}

				log.WithFields(log.Fields{
					"file":          "lib/fz/notification.go",
					"notifier_name": n.Notifier.Name,
				}).Info("sending notification")

				ctx, cancel := context.WithTimeout(context.Background(), n.Notifier.timeout())
				defer cancel()

				cmd := exec.CommandContext(ctx, n.Notifier.Command, n.Notifier.Args...)

				stdin, err := cmd.StdinPipe()
				if err != nil {
					log.WithFields(log.Fields{
						"file":          "lib/fz/notification.go",
						"notifier_name": n.Notifier.Name,
					}).Error(err)
				}

				cmd.Env = []string{
					fmt.Sprintf("PRIORITY=%d", n.Task.Priority),
					fmt.Sprintf("SUBJECT=%s", n.subject()),
					fmt.Sprintf("STATE=%s", n.Task.State()),
					fmt.Sprintf("LAST_STATE=%s", n.Task.LastState()),
				}

				log.WithFields(log.Fields{
					"file":          "lib/fz/notification.go",
					"notifier_name": n.Notifier.Name,
				}).Trace(fmt.Sprintf("writing string to stdin: %s", n.body()))

				io.WriteString(stdin, n.body())
				stdin.Close()

				err = cmd.Run()

				if ctx.Err() == context.DeadlineExceeded {
					log.WithFields(log.Fields{
						"file":          "lib/fz/notification.go",
						"notifier_name": n.Notifier.Name,
					}).Error(fmt.Sprintf("time out exceeded while executing notifier"))
				} else if err != nil {
					log.WithFields(log.Fields{
						"file":          "lib/fz/notification.go",
						"notifier_name": n.Notifier.Name,
					}).Error(err)
				}
			}
		}
	}()
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
