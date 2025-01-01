package fz

import (
	"fmt"
	"time"
)

var NotifyCh = make(chan TaskNotification, 100)
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
				if !n.GateSetOpen() {
					Logger.Debug("notification cancelled. all gatesets are closed.",
						"notifier", n.Notifier.Name,
						"task", n.Task.Name,
						"trace_id", n.TraceID,
					)
					break C
				}

				n.Task.LastNotification = time.Now()
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

// evaluate the state of the gatesets, and return true if any set is completely open.
func (n TaskNotification) GateSetOpen() bool {
	return gateSetOpen(n.Task, n.Notifier.GateSets())
}

// evaluate the state of the gatesets, and return true if any set is completely open.
func (n ErrorNotification) GateSetOpen() bool {
	return gateSetOpen(&Task{TraceID: n.TraceID}, n.Notifier.GateSets())
}

// The environment variables provided to the notifiers
func (n TaskNotification) Environment(tasks ...*Task) []string {
	if len(n.Message) == 0 {
		n.Message = "no message recieved"
	}

	v := []string{
		fmt.Sprintf("MSG=%s", n.Message),
		fmt.Sprintf("SUBJECT=%s: state is %s", n.Task.Name, n.Task.State()),
		fmt.Sprintf("TASK_DURATION_MS=%d", n.Duration.Milliseconds()),
		fmt.Sprintf("TASK_EPOCH=%d", n.Timestamp.Unix()),
		fmt.Sprintf("TASK_LAST_NOTIFICATION=%d", n.Task.LastNotification.Unix()),
		fmt.Sprintf("TASK_LAST_STATE=%s", n.Task.LastState()),
		fmt.Sprintf("TASK_NAME=%s", n.Task.Name),
		fmt.Sprintf("TASK_PRIORITY=%d", n.Task.Priority),
		fmt.Sprintf("TASK_STATE=%s", n.Task.State()),
		fmt.Sprintf("TASK_TIMEOUT_MS=%d", n.Task.TimeoutSeconds*1000),
		fmt.Sprintf("TASK_TRACE_ID=%s", n.Task.TraceID),
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
		fmt.Sprintf("TASK_TRACE_ID=%s", n.TraceID),
	}

	v = MergeEnvVars(v, n.Notifier.Envs)
	v = MergeEnvVars(v, cfg.Defaults.Envs)

	return v
}

func gateSetOpen(t *Task, gatesets [][]*Gate) bool {
	for _, gs := range gatesets {
		if GateSetOpen(t, gs...) {
			return true
		}
	}

	// no gatesets were open
	return false
}
