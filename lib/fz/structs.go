package fz

import (
	"time"

	"github.com/jonhiggs/flamingzombies/lib/config"
)

///////////////////////////////////////////////////////////////////////////////
// Constants

const VERSION = "v0.1.0"

const DEFAULT_RETRIES = 5
const DEFAULT_LOG_LEVEL = "info"
const DEFAULT_LOG_FILE = "-"
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300
const DEFAULT_PRIORITY = 5
const DEFAULT_GATE_TIMEOUT_SECONDS = 1

///////////////////////////////////////////////////////////////////////////////
// Structs

// TODO(jh) 20241231: Finish setting this up.
type Notification interface {
	GateSetOpen() bool
	Environment() []string
}

// A notification is generated upon the successful completion of any task. It
// extracts data from the task to provide to the notifier.
type TaskNotification struct {
	Duration  time.Duration
	Notifier  *config.Notifier
	Task      config.Task
	Timestamp time.Time
	Message   string
	TraceID   string
}

// An ErrorNotification are generated on error events. This are never expected
// and generally not gated. They're similar to a regular Notification, except
// their payload is an error type rather than a Task.
type ErrorNotification struct {
	Notifier *config.Notifier
	Error    error
	TraceID  string
}
