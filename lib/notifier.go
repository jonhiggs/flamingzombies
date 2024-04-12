package fz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type Notifier struct {
	Name           string
	Command        string
	Args           []string
	MinPriority    int
	TimeoutSeconds int `toml:"timeout_seconds"`
}

type Notification struct {
	Notifier *Notifier
	Subject  string
	Body     string
	Priority int
}

var NotifyCh = make(chan Notification, 100)

func ProcessNotifications() {
	go func() {
		for {
			select {
			case n := <-NotifyCh:
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
					fmt.Sprintf("PRIORITY=%d", n.Priority),
					fmt.Sprintf("SUBJECT=%s", n.Subject),
				}

				io.WriteString(stdin, n.Body+"\n")

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
