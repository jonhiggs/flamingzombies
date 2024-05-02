package fz

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Name                  string   `toml:"name"`            // friendly name
	Command               string   `toml:"command"`         // command
	Args                  []string `toml:"args"`            // command arguments
	FrequencySeconds      int      `toml:"frequency"`       // how often to run
	RetryFrequencySeconds int      `toml:"retry_frequency"` // how quickly to retry when state unknown
	TimeoutSeconds        int      `toml:"timeout"`         // how long an execution may run
	Retries               int      `toml:"retries"`         // number of retries before changing the state
	NotifierNames         []string `toml:"notifiers"`       // notifiers to trigger upon state change
	Priority              int      `toml:"priority"`        // the priority of the notifications
	ErrorBody             string   `toml:"error_body"`      // the body of the notification when entering an error state
	RecoverBody           string   `toml:"recover_body"`    // the body of the notification when recovering from an error state

	// public, but unconfigurable
	LastRun        time.Time
	LastOk         time.Time
	History        uint32 // represented in binary. Successes are high
	HistoryMask    uint32 // the bits in the history with a recorded value. Needed to understand a history of 0
	ExecutionCount int    // task was executed
	OKCount        int    // task passed
	FailCount      int    // task failed
	ErrorCount     int    // fask failed to executed

	mutex sync.Mutex // lock to ensure one task runs at a time
}

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
	// if the state is unknown, retry at the rate of RetryFrequencySeconds

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

	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	t.LastRun = time.Now()
	t.ExecutionCount++

	errorMessage, _ := io.ReadAll(stderr)

	err = cmd.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		Logger.Error("time out exceeded while executing command", "task", t.Name)
		t.ErrorCount++
		return false
	}

	if err != nil {
		if os.IsPermission(err) {
			Logger.Error(fmt.Sprint(err), "task", t.Name)
			t.ErrorCount++
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
		return false
	case 124: // unknown status due to timeout
		return false
	case 0:
		t.LastOk = time.Now()
		t.OKCount++
		t.RecordStatus(true)
	default:
		t.FailCount++
		t.RecordStatus(false)
	}

	// raising notifications
	if t.State() != STATE_UNKNOWN {
		for _, name := range t.NotifierNames {
			for i, n := range config.Notifiers {
				if n.Name == name {
					Logger.Debug("raising notification", "task", t.Name, "last_state", t.LastState, "new_state", t.State())
					NotifyCh <- Notification{&config.Notifiers[i], t}
				}
			}
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
	return time.Duration(t.TimeoutSeconds) * time.Second
}

func (t Task) retryFrequency() time.Duration {
	return time.Duration(t.RetryFrequencySeconds) * time.Second
}

func (t Task) notifiers() []*Notifier {
	var not []*Notifier
	for _, nName := range t.NotifierNames {
		found := false
		for i, _ := range config.Notifiers {
			if nName == config.Notifiers[i].Name {
				not = append(not, &config.Notifiers[i])
				found = true
			}
		}

		if !found {
			panic(fmt.Sprintf("unknown notifier '%s'", nName))
		}
	}

	return not
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

func (t Task) validate() error {
	if _, err := os.Stat(t.Command); os.IsNotExist(err) {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, t.Command)); os.IsNotExist(err) {
			return fmt.Errorf("task command not found")
		}
	}

	return nil
}
