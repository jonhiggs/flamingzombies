package fz

import (
	"sync"
	"time"
)

type Config struct {
	Defaults       ConfigDefaults
	Directory      string     `toml:"directory"`
	ErrorNotifiers []Notifier `toml:"error_notifier"`
	Gates          []Gate     `toml:"gate"`
	ListenAddress  string     `toml:"listen_address"`
	LogLevel       string     `toml:"log_level"`
	Notifiers      []Notifier `toml:"notifier"`
	StatsdHost     string     `toml:"statsd_host"`
	StatsdPrefix   string     `toml:"statsd_prefix"`
	Tasks          []Task     `toml:"task"`
}

type ConfigDefaults struct {
	FrequencySeconds      int        `toml:"frequency"`
	NotifierNames         []string   `toml:"notifiers"`
	Priority              int        `toml:"priority"`
	Retries               int        `toml:"retries"`
	RetryFrequencySeconds int        `toml:"retry_frequency"`
	TaskEnvs              [][]string `toml:"task_envs`
	TimeoutSeconds        int        `toml:"timeout"` // better to put the timeout into the command
}

type NotifierKind int

const (
	TaskNotifierKind NotifierKind = iota
	ErrorNotifierKind
)

type Notifier struct {
	GateSets       [][]string `toml:"gates"`
	TimeoutSeconds int        `toml:"timeout"`
	Args           []string   `toml:"args"`
	Command        string     `toml:"command"`
	Name           string     `toml:"name"`

	kind  NotifierKind
	gates [][]*Gate
}

type Task struct {
	Name                  string     `toml:"name"`            // friendly name
	Command               string     `toml:"command"`         // command
	Args                  []string   `toml:"args"`            // command arguments
	FrequencySeconds      int        `toml:"frequency"`       // how often to run
	RetryFrequencySeconds int        `toml:"retry_frequency"` // how quickly to retry when state unknown
	TimeoutSeconds        int        `toml:"timeout"`         // how long an execution may run
	Retries               int        `toml:"retries"`         // number of retries before changing the state
	NotifierNames         []string   `toml:"notifiers"`       // notifiers to trigger upon state change
	Priority              int        `toml:"priority"`        // the priority of the notifications
	Envs                  [][]string `toml:"envs`             // environment variables supplied to task
	ErrorBody             string     `toml:"error_body"`      // the body of the notification when entering an error state
	RecoverBody           string     `toml:"recover_body"`    // the body of the notification when recovering from an error state

	// public, but not configurable
	ErrorCount       int       // task failed to executed
	ExecutionCount   int       // task was executed
	FailCount        int       // task failed
	History          uint32    // represented in binary. Successes are high
	HistoryMask      uint32    // the bits in the history with a recorded value. Needed to understand a history of 0
	LastFail         time.Time // the time of the last failed execution
	LastOk           time.Time // the time of the last successful execution
	LastResultOutput string    // the result output of the last execution
	LastRun          time.Time // the time of the last execution
	OKCount          int       // task passed

	mutex             sync.Mutex  // lock to ensure one task runs at a time
	lastNotifications []time.Time // times that each notifier was last executed
}

type Gate struct {
	Name    string   `toml:"name"`    // friendly name
	Command string   `toml:"command"` // command
	Args    []string `toml:"args"`    // command arguments
}

type Notification struct {
	Notifier *Notifier
	Task     *Task
}
