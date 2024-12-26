package fz

import (
	"errors"
	"sync"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// Constants

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300
const DEFAULT_PRIORITY = 5
const GRACE_TIME = time.Duration(500) * time.Millisecond

///////////////////////////////////////////////////////////////////////////////
// Errors

var ErrCommandNotExist = errors.New("command does not exist")
var ErrInvalidName = errors.New("charactors must be alphanumeric or underscore")
var ErrNotExist = errors.New("does not exist")
var ErrLessThan1 = errors.New("cannot be less than 1")
var ErrTimeoutSlowerThanRetry = errors.New("timeout must not be longer than the retry interval")
var ErrGreaterThan99 = errors.New("cannot be greater than 99")

///////////////////////////////////////////////////////////////////////////////
// Structs

type Config struct {
	Defaults  ConfigDefaults `toml:"defaults"`
	Directory string         `toml:"directory"`
	Gates     []Gate         `toml:"gate"`
	LogFile   string         `toml:"log_file"`
	LogLevel  string         `toml:"log_level"`
	Notifiers []Notifier     `toml:"notifier"`
	Tasks     []Task         `toml:"task"`
}

type ConfigDefaults struct {
	FrequencySeconds      int        `toml:"frequency"`
	NotifierNames         []string   `toml:"notifiers"`
	ErrorNotifierNames    []string   `toml:"error_notifiers"`
	Priority              int        `toml:"priority"`
	Retries               int        `toml:"retries"`
	RetryFrequencySeconds int        `toml:"retry_frequency"`
	Envs                  [][]string `toml:"envs`
	TimeoutSeconds        int        `toml:"timeout"` // better to put the timeout into the command
}

// A Notifier script is capable of emitting an event to an external service.
type Notifier struct {
	GateSets       [][]string `toml:"gates"`
	TimeoutSeconds int        `toml:"timeout"`
	Args           []string   `toml:"args"`
	Command        string     `toml:"command"`
	Name           string     `toml:"name"`
	Envs           [][]string `toml:"envs`
}

// A Task is a command that is executed on a schedule. The struct contains the
// static configuration of the task which is read from the configuration file,
// and it's metadata and history which are generated over the course of the
// daemons lifecycle.
type Task struct {
	Name                  string     `toml:"name"`            // friendly name
	Command               string     `toml:"command"`         // command
	Args                  []string   `toml:"args"`            // command arguments
	FrequencySeconds      int        `toml:"frequency"`       // how often to run
	RetryFrequencySeconds int        `toml:"retry_frequency"` // how quickly to retry when state unknown
	TimeoutSeconds        int        `toml:"timeout"`         // how long an execution may run
	Retries               int        `toml:"retries"`         // number of retries before changing the state
	NotifierNames         []string   `toml:"notifiers"`       // notifiers to trigger upon state change
	ErrorNotifierNames    []string   `toml:"error_notifiers"` // notifiers to trigger upon state change
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

// The gate is the control mechanism to governs whether a Notifier executes.
type Gate struct {
	Args    []string   `toml:"args"`    // command arguments
	Command string     `toml:"command"` // command
	Envs    [][]string `toml:"envs`     // environment variables
	Name    string     `toml:"name"`    // friendly name
}

// A notification is generated upon the successful completion of any task.
type Notification struct {
	Notifier *Notifier
	Task     *Task
}

// An ErrorNotification are generated on error events. This are never expected
// and generally not gated.
type ErrorNotification struct {
	Notifier *Notifier
	Error    error
}
