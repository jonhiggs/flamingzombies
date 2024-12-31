package fz

import (
	"fmt"
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
				n.Notifier.Execute(n.TraceID, n.Environment(), false)

			case n := <-NotifyCh:
				_, ok := n.gateEvaluate()
				if !ok {
					Logger.Debug("notification cancelled. all gates are closed.",
						"notifier", n.Notifier.Name,
						"task", n.Task.Name,
						"trace_id", n.TraceID,
					)
					break C
				}

				Logger.Info("sending notification",
					"notifier", n.Notifier.Name,
					"task", n.Task.Name,
					"trace_id", n.TraceID,
				)
				n.Notifier.Execute(n.TraceID, n.Environment(n.Task), true)
			}
		}
	}()
}

// evaluate the state of the gatesets, and return true if the gates are open.
func (n Notification) gateEvaluate() ([]*Gate, bool) {
	openGates := []*Gate{}
	closedGates := []*Gate{}
X:
	for gsi, gs := range cfg.GetNotifierGateSets(n.Notifier.Name) {
		openGates = []*Gate{} // ignore the gates from prior gateset

		for _, g := range gs {
			isOpen, r := g.IsOpen(n.Task)
			Logger.Debug("checking gate",
				"name", g.Name,
				"stdout", string(r.StdoutBytes),
				"stderr", string(r.StderrBytes),
				"exit_code", r.ExitCode,
				"trace_id", n.TraceID,
			)
			if !isOpen {
				Logger.Debug("gate is closed",
					"name", g.Name,
					"notifier", n.Notifier.Name,
					"task", n.Task.Name,
					"trace_id", n.TraceID,
				)
				closedGates = append(closedGates, g)
				continue X
			}

			openGates = append(openGates, g)
			Logger.Debug("gate is open",
				"name", g.Name,
				"notifier", n.Notifier.Name,
				"task", n.Task.Name,
				"trace_id", n.TraceID,
			)
		}
		Logger.Debug("gateset is open",
			"gateset", gsi,
			"trace_id", n.TraceID,
		)
		return openGates, true
	}

	return openGates, (len(closedGates) == 0)
}

// The environment variables provided to the notifiers
func (n Notification) Environment(tasks ...*Task) []string {
	v := []string{
		fmt.Sprintf("MSG=%s", n.Task.LastResultOutput),
		fmt.Sprintf("SUBJECT=%s: state is %s", n.Task.Name, n.Task.State()),
		fmt.Sprintf("TASK_DURATION_MS=%d", n.Duration.Milliseconds()),
		fmt.Sprintf("TASK_EPOCH=%d", n.Timestamp.Unix()),
		fmt.Sprintf("TASK_LAST_STATE=%s", n.Task.LastState()),
		fmt.Sprintf("TASK_NAME=%s", n.Task.Name),
		fmt.Sprintf("TASK_PRIORITY=%d", n.Task.Priority),
		fmt.Sprintf("TASK_STATE=%s", n.Task.State()),
		fmt.Sprintf("TASK_TIMEOUT_MS=%d", n.Task.TimeoutSeconds*1000),
	}

	for _, t := range tasks {
		v = MergeEnvVars(v, t.Envs)
	}

	v = MergeEnvVars(v, n.Notifier.Envs)
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}

// The environment variables provided to the error_notifiers
func (n ErrorNotification) Environment() []string {
	v := []string{
		fmt.Sprintf("MSG=%s", n.Error),
		fmt.Sprintf("SUBJECT=%s", "fz experienced a critical error"),
	}

	v = MergeEnvVars(v, n.Notifier.Envs)
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}
