package fz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
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
				openGates, ok := n.gateState()
				if !ok {
					Logger.Debug("notification canceled due to a closed gate", "notifier", n.Notifier.Name)
					break C
				}

				var openGatesNames []string
				for _, g := range openGates {
					openGatesNames = append(openGatesNames, g.Name)
				}

				Logger.Info("sending notification", "notifier", n.Notifier.Name)
				n.Task.SetLastNotification(n.Notifier.Name, time.Now())

				ctx, cancel := context.WithTimeout(context.Background(), n.Notifier.timeout())
				defer cancel()

				cmd := exec.CommandContext(ctx, n.Notifier.Command, n.Notifier.Args...)

				stdin, err := cmd.StdinPipe()
				if err != nil {
					Logger.Error(fmt.Sprint(err), "notifier", n.Notifier.Name)
				}

				cmd.Dir = config.Directory
				cmd.Env = []string{
					fmt.Sprintf("LAST_STATE=%s", n.Task.LastState()),
					fmt.Sprintf("NAME=%s", n.Task.Name),
					fmt.Sprintf("OPEN_GATES=%s", strings.Join(openGatesNames, ",")),
					fmt.Sprintf("PRIORITY=%d", n.Task.Priority),
					fmt.Sprintf("STATE=%s", n.Task.State()),
					fmt.Sprintf("RESULT_OUTPUT=%s", n.Task.LastResultOutput),
				}

				io.WriteString(stdin, n.body())
				stdin.Close()

				stderr, _ := cmd.StderrPipe()

				err = cmd.Start()
				if err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						Logger.Error(fmt.Sprintf("time out exceeded while executing notifier"), "notifier", n.Notifier.Name)
					} else {
						panic(err)
					}
				}

				errorMessage, _ := io.ReadAll(stderr)

				err = cmd.Wait()

				if ctx.Err() == context.DeadlineExceeded {
					Logger.Error(fmt.Sprintf("time out exceeded while executing notifier"), "notifier", n.Notifier.Name)
				} else if err != nil {
					exiterr, _ := err.(*exec.ExitError)
					exitCode := exiterr.ExitCode()

					Logger.Error(fmt.Sprintf("command returned stderr: %s", errorMessage), "notifier", n.Notifier.Name, "exit_code", exitCode)
				}
			}
		}
	}()
}

// check the state of all configured gates.
func (n Notification) gateState() ([]*Gate, bool) {
	openGates := []*Gate{}
X:
	for gsi, gs := range n.Notifier.Gates() {
		openGates = []*Gate{} // ignore the gates from prior gateset

		for _, g := range gs {
			if g.IsOpen(n.Task, n.Notifier) == false {
				Logger.Debug("gate is closed", "gate", g.Name)
				continue X
			}

			openGates = append(openGates, g)
			Logger.Debug("gate is open", "gate", g.Name)
		}
		Logger.Debug("gateset is open", "gateset", gsi)
		return openGates, true
	}

	return openGates, false
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
