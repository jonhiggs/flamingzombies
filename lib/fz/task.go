package fz

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

const GRACE_TIME = time.Duration(500) * time.Millisecond

func (t Task) Hash() uint32 {
	// To help with testing, return hash of zero when there isn't a command or
	// any arguments.
	if t.Command == "" && len(t.Args) == 0 {
		return uint32(0)
	}

	s := t.Command
	for _, a := range t.Args {
		s += a
	}

	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// how often to run
func (t Task) Frequency() time.Duration {
	if t.FrequencySeconds == 0 {
		return time.Duration(DEFAULT_FREQUENCY_SECONDS) * time.Second
	}

	return time.Duration(t.FrequencySeconds) * time.Second
}

func (t Task) Ready(ts time.Time) bool {
	// the hash is used to spread the checks across time.
	// while the state is unknown, retry at the rate of RetryFrequencySeconds

	if t.State() == STATE_UNKNOWN {
		return (uint32(ts.Unix())+t.Hash())%uint32(t.RetryFrequencySeconds) == 0
	}

	return (uint32(ts.Unix())+t.Hash())%uint32(t.FrequencySeconds) == 0
}

func (t *Task) Run() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	Logger.Info("executing task", "task", t.Name)

	ctx, cancel := context.WithTimeout(context.Background(), t.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command, t.ExpandArgs()...)
	cmd.Dir = config.Directory

	cmd.Env = []string{
		fmt.Sprintf("TIMEOUT=%d", t.TimeoutSeconds),
	}

	for _, v := range config.Defaults.Envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", v[0], v[1]))
	}

	for _, v := range t.Envs {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", v[0], v[1]))
	}

	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	//startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	t.LastRun = time.Now()
	t.ExecutionCount++

	errorMessage, _ := io.ReadAll(stderr)
	stdoutBytes, _ := io.ReadAll(stdout)
	t.LastResultOutput = strings.TrimSuffix(string(stdoutBytes), "\n")

	err = cmd.Wait()
	//t.DurationMetric(time.Now().Sub(startTime))

	if ctx.Err() == context.DeadlineExceeded {
		Logger.Error("time out exceeded while executing command", "task", t.Name)
		t.ErrorCount++
		//t.IncMetric("timeout")

		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			Logger.Error(fmt.Sprint(err), "task", t.Name)
			t.ErrorCount++
			//t.IncMetric("error")

			return false
		}
	}

	var exitCode int
	if err != nil {
		exiterr, _ := err.(*exec.ExitError)
		exitCode = exiterr.ExitCode()
	} else {
		exitCode = 0
	}

	Logger.Debug(fmt.Sprintf("command returned stderr: %s", errorMessage), "task", t.Name)

	switch exitCode {
	case 3: // unknown status
		//t.IncMetric("unknown")
		return false
	case 124: // unknown status due to timeout
		//t.IncMetric("unknown")
		return false
	case 0:
		t.OKCount++
		//t.IncMetric("ok")
		t.RecordStatus(true)
	default:
		t.FailCount++
		//t.IncMetric("fail")
		t.RecordStatus(false)
	}

	// raising notifications
	if t.State() != STATE_UNKNOWN {
		for _, n := range t.notifiers() {
			Logger.Debug("raising notification", "task", t.Name, "last_state", t.LastState(), "new_state", t.State())
			NotifyCh <- Notification{n, t}
		}
	}
	return true
}

func (t *Task) RecordStatus(b bool) {
	Logger.Debug(fmt.Sprintf("recording measurement %v", b), "task", t.Name)

	t.History = t.History << 1
	if b {
		t.History += 1
	}

	t.HistoryMask = t.HistoryMask << 1
	t.HistoryMask += 1

	switch t.State() {
	case STATE_OK:
		t.LastOk = time.Now()
	case STATE_FAIL:
		t.LastFail = time.Now()
	}
}

// extract the current state from the history
func (t Task) State() State {
	// if there aren't enough measurements, return STATE_UNKNOWN
	if t.retryMask() > t.HistoryMask {
		return STATE_UNKNOWN
	}

	v := t.History & t.retryMask()

	if v == 0 {
		return STATE_FAIL
	}

	if v == t.retryMask() {
		return STATE_OK
	}

	return STATE_UNKNOWN
}

// step back though the data to find the previous state
func (t Task) LastState() State {
	h := t.History >> t.Retries
	m := t.HistoryMask >> t.Retries

	mask := t.retryMask()

	for mask <= m {
		if h&mask == mask {
			return STATE_OK
		}

		if h&mask == 0 {
			return STATE_FAIL
		}

		h = h >> 1
		m = m >> 1
	}

	return STATE_UNKNOWN
}

// if the state changed
func (t Task) StateChanged() bool {
	// if state is unknown, then we can't make an assessment.
	if t.State() == STATE_UNKNOWN {
		return false
	}

	// shift back to the last record. if we had the data to raise an alert,
	// then assume we did.
	l := (t.History >> 1) & t.retryMask()
	if l == (t.History & t.retryMask()) {
		return false
	}

	if t.LastState() == STATE_UNKNOWN {
		return false
	}

	return t.State() != t.LastState()
}

func (t Task) retryMask() uint32 {
	var m uint32
	for i := 0; i < t.Retries; i++ {
		m = m << 1
		m += 1
	}

	return m
}

func (t Task) timeout() time.Duration {
	return GRACE_TIME + time.Duration(t.TimeoutSeconds)*time.Second
}

func (t Task) retryFrequency() time.Duration {
	return time.Duration(t.RetryFrequencySeconds) * time.Second
}

func (t Task) notifiers() []*Notifier {
	var ns []*Notifier
	for _, nName := range t.NotifierNames {
		for i, _ := range config.Notifiers {
			if nName == config.Notifiers[i].Name {
				n := &config.Notifiers[i]
				n.kind = TaskNotifierKind
				ns = append(ns, n)
			}
		}
	}

	return ns
}

// return the arguments after interpolating the values
func (t Task) ExpandArgs() []string {
	var newArgs []string

	for _, a := range t.Args {
		a = strings.ReplaceAll(a, "%{TIMEOUT_SECONDS}", fmt.Sprintf("%d", t.TimeoutSeconds))
		newArgs = append(newArgs, a)
	}

	return newArgs
}

func (t Task) NotifierIndex(name string) (int, error) {
	for i, n := range t.NotifierNames {
		if n == name {
			return i, nil
		}
	}
	return -1, fmt.Errorf("unknown notifier name")
}

// get the last notification of all notifiers.
func (t Task) LastNotification() time.Time {
	var ts time.Time

	for i, n := range t.lastNotifications {
		if i == 0 {
			ts = n
		} else {
			if n.Unix() > ts.Unix() {
				ts = n
			}
		}
	}

	return ts
}

func (t Task) GetLastNotification(name string) time.Time {
	i, err := t.NotifierIndex(name)
	if err != nil {
		panic(err)
	}

	if len(t.lastNotifications) != len(t.NotifierNames) {
		return DAEMON_START_TIME
	}

	return t.lastNotifications[i]
}

func (t *Task) SetLastNotification(name string, ts time.Time) error {
	if len(t.lastNotifications) != len(t.NotifierNames) {
		t.lastNotifications = make([]time.Time, len(t.NotifierNames))
	}

	i, err := t.NotifierIndex(name)
	if err != nil {
		return err
	}

	t.lastNotifications[i] = ts

	return nil
}
