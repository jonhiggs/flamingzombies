package fz

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

var NotifyCh = make(chan Notification, 100)
var ErrorNotifyCh = make(chan ErrorNotification, 100)

func ProcessNotifications() {
	go func() {
		for {
		C:
			select {
			case n := <-ErrorNotifyCh:
				Logger.Info("sending error notification", "notifier", n.Notifier.Name)

				ctx, cancel := context.WithTimeout(context.Background(), n.Notifier.Timeout())
				defer cancel()

				cmd := exec.CommandContext(ctx, n.Notifier.Command, n.Notifier.Args...)

				stdin, err := cmd.StdinPipe()
				if err != nil {
					Logger.Error(fmt.Sprint(err), "notifier", n.Notifier.Name)
				}

				cmd.Dir = cfg.Directory

				//io.WriteString(stdin, n.body())
				stdin.Close()

				stderr, _ := cmd.StderrPipe()

				//startTime := time.Now()
				err = cmd.Start()
				if err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						Logger.Error(fmt.Sprintf("time out exceeded while executing notifier"), "notifier", n.Notifier.Name)
					} else {
						// TODO: Handle this error
						panic(err)
					}
				}

				errorMessage, _ := io.ReadAll(stderr)

				err = cmd.Wait()
				//n.DurationMetric(time.Now().Sub(startTime))

				if ctx.Err() == context.DeadlineExceeded {
					Logger.Error(fmt.Sprintf("time out exceeded while executing error notifier"), "notifier", n.Notifier.Name)
				} else if err != nil {
					exiterr, _ := err.(*exec.ExitError)
					exitCode := exiterr.ExitCode()

					Logger.Error(fmt.Sprintf("command returned stderr: %s", errorMessage), "notifier", n.Notifier.Name, "exit_code", exitCode)
				}

			case n := <-NotifyCh:
				_, ok := n.gateState()
				if !ok {
					Logger.Debug("notification cancelled due to a closed gate",
						"notifier", n.Notifier.Name,
						"task", n.Task.Name,
					)
					break C
				}

				Logger.Info("sending notification", "notifier", n.Notifier.Name)

				ctx, cancel := context.WithTimeout(context.Background(), n.Notifier.Timeout())
				defer cancel()

				cmd := exec.CommandContext(ctx, n.Notifier.Command, n.Notifier.Args...)

				stdin, err := cmd.StdinPipe()
				if err != nil {
					Logger.Error(fmt.Sprint(err), "notifier", n.Notifier.Name)
				}

				cmd.Dir = cfg.Directory
				cmd.Env = n.environment()

				io.WriteString(stdin, n.body())
				stdin.Close()

				stderr, _ := cmd.StderrPipe()

				//startTime := time.Now()
				err = cmd.Start()
				if err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						Logger.Error(fmt.Sprintf("time out exceeded while executing notifier"), "notifier", n.Notifier.Name)
						//n.IncMetric("timeout")
					} else {
						// TODO: handle this error
						panic(err)
					}
				}

				errorMessage, _ := io.ReadAll(stderr)

				err = cmd.Wait()
				//n.DurationMetric(time.Now().Sub(startTime))

				if ctx.Err() == context.DeadlineExceeded {
					Logger.Error(fmt.Sprintf("time out exceeded while executing notifier"), "notifier", n.Notifier.Name)
					//n.IncMetric("timeout")
				} else if err != nil {
					exiterr, _ := err.(*exec.ExitError)
					exitCode := exiterr.ExitCode()

					Logger.Error(fmt.Sprintf("command returned stderr: %s", errorMessage), "notifier", n.Notifier.Name, "exit_code", exitCode)
					//n.IncMetric("error")
				}
			}
		}
	}()
}

// check the state of all configured gates.
func (n Notification) gateState() ([]*Gate, bool) {
	openGates := []*Gate{}
	closedGates := []*Gate{}
X:
	for gsi, gs := range cfg.GetNotifierGateSets(n.Notifier.Name) {
		openGates = []*Gate{} // ignore the gates from prior gateset

		for _, g := range gs {
			if g.IsOpen(n.Task, n.Notifier) == false {
				Logger.Debug("gate is closed", "gate", g.Name)
				closedGates = append(closedGates, g)
				continue X
			}

			openGates = append(openGates, g)
			Logger.Debug("gate is open",
				"name", g.Name,
				"task", n.Task.Name,
			)
		}
		Logger.Debug("gateset is open", "gateset", gsi)
		return openGates, true
	}

	return openGates, (len(closedGates) == 0)
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

// The environment variables provided to the notifier
func (n Notification) environment() []string {
	v := []string{
		fmt.Sprintf("DURATION_MS=%d", n.Duration.Milliseconds()),
		fmt.Sprintf("EPOCH=%d", n.Timestamp.Unix()),
		fmt.Sprintf("PRIORITY=%d", n.Task.Priority),
		fmt.Sprintf("LAST_STATE=%s", n.Task.LastState()),
		fmt.Sprintf("NAME=%s", n.Task.Name),
		fmt.Sprintf("OUTPUT=%s", n.Task.LastResultOutput),
		fmt.Sprintf("STATE=%s", n.Task.State()),
		fmt.Sprintf("TIMEOUT_MS=%d", n.Task.TimeoutSeconds*1000),
	}

	for _, e := range n.Notifier.Envs {
		v = append(v, e)
	}

	return v
}

// TODO: Provide the data to the notifier so it can publish the metrics to statsd or elsewhere.
//func (n Notification) IncMetric(x string) {
//	StatsdClient.Inc(
//		fmt.Sprintf("notifier.%s", x), 1, 1.0,
//		statsd.Tag{"host", Hostname},
//		statsd.Tag{"name", n.Notifier.Name},
//	)
//}
//
//func (n Notification) DurationMetric(d time.Duration) {
//	StatsdClient.TimingDuration(
//		"notifier.duration", d, 1.0,
//		statsd.Tag{"host", Hostname},
//		statsd.Tag{"name", n.Notifier.Name},
//	)
//
//	StatsdClient.Gauge(
//		"notifier.timeoutquota.percent", int64(float64(d)/float64(n.Notifier.timeout())*100), 1.0,
//		statsd.Tag{"host", Hostname},
//		statsd.Tag{"name", n.Notifier.Name},
//	)
//}
