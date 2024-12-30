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
				n.Notifier.Execute(n.Environment(), false)

			case n := <-NotifyCh:
				_, ok := n.gateEvaluate()
				if !ok {
					Logger.Debug("notification cancelled. all gates are closed.",
						"notifier", n.Notifier.Name,
						"task", n.Task.Name,
					)
					break C
				}

				Logger.Info("sending notification", "notifier", n.Notifier.Name)
				n.Notifier.Execute(n.Environment(n.Task), true)
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
			if g.Execute(n.Task) == false {
				Logger.Debug("gate is closed",
					"gate", g.Name,
				)
				closedGates = append(closedGates, g)
				continue X
			}

			openGates = append(openGates, g)
			Logger.Debug("gate is open",
				"name", g.Name,
				"task", n.Task.Name,
			)
		}
		Logger.Debug("gateset is open",
			"gateset", gsi,
		)
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
