package fz

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Create a checksum of a tasks configuration. The hash is used for a
// consistent execution offset. Offsetting the execution prevents the time that
// tasks are executed from clustering around each other.
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

func (t *Task) Run() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	Logger.Info("executing task", "task", t.Name)

	ctx, cancel := context.WithTimeout(context.Background(), t.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, t.Command, t.Args...)
	cmd.Dir = cfg.Directory
	cmd.Env = t.Environment()

	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	//startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		Logger.Error(fmt.Sprintf("%s", err))
		// TODO: emit an ErrorNotification
		return
	}

	t.LastRun = time.Now()

	errorMessage, _ := io.ReadAll(stderr)
	stdoutBytes, _ := io.ReadAll(stdout)
	t.LastResultOutput = strings.TrimSuffix(string(stdoutBytes), "\n")

	err = cmd.Wait()
	//t.DurationMetric(time.Now().Sub(startTime))

	if ctx.Err() == context.DeadlineExceeded {
		Logger.Error("time out exceeded while executing command", "task", t.Name)
		//t.IncMetric("timeout")

		return
	}

	if err != nil {
		if os.IsPermission(err) {
			Logger.Error(fmt.Sprint(err), "task", t.Name)
			//t.IncMetric("error")

			return
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
		return
	case 124: // unknown status due to timeout
		//t.IncMetric("unknown")
		return
	case 0:
		//t.IncMetric("ok")
		t.RecordStatus(true)
	default:
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
	return
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

// Check that the task is in a valid state.
func (t Task) Validate() error {
	re := regexp.MustCompile(`^[a-z0-9_]+$`)
	if !re.Match([]byte(t.Name)) {
		return fmt.Errorf("name '%s': %w", t.Name, ErrInvalidName)
	}

	// The command string cannot be blank. When loading the configuration, a
	// better test is performed to make sure that the file actually exists.
	if len(t.Command) < 1 {
		return fmt.Errorf("command '%s': %w", t.Command, ErrCommandNotExist)
	}

	if t.FrequencySeconds < 1 {
		return fmt.Errorf("freqency '%d': %w", t.FrequencySeconds, ErrLessThan1)
	}

	if t.RetryFrequencySeconds < 1 {
		return fmt.Errorf("retry_freqency '%d': %w", t.RetryFrequencySeconds, ErrLessThan1)
	}

	if t.TimeoutSeconds < 1 {
		return fmt.Errorf("timeout_seconds '%d': %w", t.TimeoutSeconds, ErrLessThan1)
	}

	if t.TimeoutSeconds > t.RetryFrequencySeconds {
		return fmt.Errorf("timeout_seconds '%d': %w", t.RetryFrequencySeconds, ErrTimeoutSlowerThanRetry)
	}

	if t.RetryFrequencySeconds > t.FrequencySeconds {
		return fmt.Errorf("retry_requency '%d': %w", t.RetryFrequencySeconds, ErrRetriesSlowerThanFrequency)
	}

	if t.Priority < 1 {
		return fmt.Errorf("priority '%d': %w", t.Priority, ErrLessThan1)
	}
	if t.Priority > 99 {
		return fmt.Errorf("priority '%d': %w", t.Priority, ErrGreaterThan99)
	}

	return nil
}

// return a list of envs that are placed into the environment when task is ran
func (t Task) Environment() []string {
	v := []string{
		fmt.Sprintf("FREQUENCY=%d", t.FrequencySeconds),
		fmt.Sprintf("HISTORY=%d", t.History),
		fmt.Sprintf("HISTORY_MASK=%d", t.HistoryMask),
		fmt.Sprintf("LAST_FAIL=%d", envEpoch(t.LastFail)),
		fmt.Sprintf("LAST_OK=%d", envEpoch(t.LastOk)),
		fmt.Sprintf("LAST_STATE=%s", t.LastState()),
		fmt.Sprintf("PRIORITY=%d", t.Priority),
		fmt.Sprintf("STATE=%s", t.State()),
		fmt.Sprintf("STATE_CHANGED=%v", t.StateChanged()),
		fmt.Sprintf("TASK_COMMAND=%s", t.Command),
		fmt.Sprintf("TIMEOUT=%d", t.TimeoutSeconds),
	}

	for _, e := range t.Envs {
		v = append(v, e)
	}

	return v
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

func (t Task) notifiers() []*Notifier {
	var ns []*Notifier
	for _, nName := range t.NotifierNames {
		for i, _ := range cfg.Notifiers {
			if nName == cfg.Notifiers[i].Name {
				n := &cfg.Notifiers[i]
				ns = append(ns, n)
			}
		}
	}

	return ns
}

// Return a UNIX epoch that's > 0
func envEpoch(t time.Time) int {
	e := t.Unix()

	if e < 0 {
		return 0
	}

	return int(e)
}
