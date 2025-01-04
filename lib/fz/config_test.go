package fz

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var workDir string

func init() {
	workDir, _ = os.Getwd()
	for !strings.HasSuffix(workDir, "flamingzombies") {
		workDir = filepath.Dir(workDir)
	}
}

func TestConfig(t *testing.T) {
	ReadConfig(
		fmt.Sprintf("%s/example_config.toml", workDir),
		fmt.Sprintf("%s/libexec", workDir),
		DEFAULT_LOG_FILE,
		DEFAULT_LOG_LEVEL,
	)

	wantLogFile := "-"
	wantLogLevel := "info"
	got := cfg

	if !strings.HasSuffix(got.Directory, "/libexec") {
		t.Errorf("expected '%s' to have '/libexec' suffix", got.Directory)
	}

	if got.LogFile != wantLogFile {
		t.Errorf("got %s, want %s", got.LogFile, wantLogFile)
	}

	if got.LogLevel != wantLogLevel {
		t.Errorf("got %s, want %s", got.LogLevel, wantLogLevel)
	}

	if len(got.Tasks) != 1 {
		t.Errorf("got %d, want %d", len(got.Tasks), 1)
	}

	if len(got.Notifiers) != 3 {
		t.Errorf("got %d, want %d", len(got.Notifiers), 3)
	}

	if len(got.Gates) != 6 {
		t.Errorf("got %d, want %d", len(got.Gates), 6)
	}

	gotTaskFlappy := got.GetTaskByName("flappy")
	if gotTaskFlappy.Name != "flappy" {
		t.Errorf("got %s, want %s", gotTaskFlappy.Name, "flappy")
	}

	gotGateSetsLogger := got.GetNotifierByName("logger").GateSets()
	if len(gotGateSetsLogger) != 2 {
		t.Errorf("got %d, want %d", len(gotGateSetsLogger), 2)
	}
	if len(gotGateSetsLogger[0]) != 3 {
		t.Errorf("got %d, want %d", len(gotGateSetsLogger[0]), 3)
	}
	if len(gotGateSetsLogger[1]) != 3 {
		t.Errorf("got %d, want %d", len(gotGateSetsLogger[1]), 3)
	}
	if gotGateSetsLogger[0][0].Name != "is_not_unknown" {
		t.Errorf("got %s, want %s", gotGateSetsLogger[0][0].Name, "is_not_unknown")
	}
	if gotGateSetsLogger[0][1].Name != "to_failed" {
		t.Errorf("got %s, want %s", gotGateSetsLogger[0][1].Name, "to_failed")
	}

	gotGateSetsStatsd := got.GetNotifierByName("statsd").GateSets()
	if len(gotGateSetsStatsd) != 0 {
		t.Errorf("got %d, want %d", len(gotGateSetsStatsd), 0)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestConfigDefaults(t *testing.T) {
	ReadConfig(
		fmt.Sprintf("%s/example_config.toml", workDir),
		fmt.Sprintf("%s/libexec", workDir),
		DEFAULT_LOG_FILE,
		DEFAULT_LOG_LEVEL,
	)

	want := ConfigDefaults{
		Retries:            5,
		TimeoutSeconds:     1,
		NotifierNames:      []string{"logger", "statsd"},
		ErrorNotifierNames: []string{"error_emailer"},
		Priority:           3,
		FrequencySeconds:   300,
		Envs: []string{
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
			"EMAIL_FROM=fz@example",
		},
	}
	got := cfg.Defaults

	if got.Retries != want.Retries {
		t.Errorf("got %d, want %d", got.Retries, want.Retries)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
	}

	if fmt.Sprintf("%s", got.NotifierNames) != fmt.Sprintf("%s", want.NotifierNames) {
		t.Errorf("got %v, want %v", got.NotifierNames, want.NotifierNames)
	}

	if fmt.Sprintf("%s", got.ErrorNotifierNames) != fmt.Sprintf("%s", want.ErrorNotifierNames) {
		t.Errorf("got %v, want %v", got.ErrorNotifierNames, want.ErrorNotifierNames)
	}

	if got.Priority != want.Priority {
		t.Errorf("got %d, want %d", got.Priority, want.Priority)
	}

	if fmt.Sprintf("%v", got.Envs) != fmt.Sprintf("%s", want.Envs) {
		t.Errorf("got %v, want %v", got.Envs, want.Envs)
	}

	if got.FrequencySeconds != want.FrequencySeconds {
		t.Errorf("got %d, want %d", got.FrequencySeconds, want.FrequencySeconds)
	}
}

func TestConfigTaskFlappy(t *testing.T) {
	ReadConfig(
		fmt.Sprintf("%s/example_config.toml", workDir),
		fmt.Sprintf("%s/libexec", workDir),
		DEFAULT_LOG_FILE,
		DEFAULT_LOG_LEVEL,
	)

	want := Task{
		Name:             "flappy",
		Command:          "task/flappy",
		FrequencySeconds: 20,
		LastNotification: time.Unix(0, 0),
	}

	wantEnvironment := []string{
		"TASK_COMMAND=task/flappy",
		"TASK_DESCRIPTION=an example task that flaps betweeen ok and fail every twenty seconds.",
		"TASK_FREQUENCY=20",
		"TASK_HISTORY=2",
		"TASK_HISTORY_MASK=0",
		"TASK_LAST_FAIL=0",
		"TASK_LAST_NOTIFICATION=0",
		"TASK_LAST_OK=0",
		"TASK_LAST_STATE=unknown",
		"TASK_NAME=flappy",
		"TASK_PRIORITY=3",
		"TASK_STATE=unknown",
		"TASK_STATE_CHANGED=false",
		"TASK_TIMEOUT=1",
		"TASK_TRACE_ID=",
		"SNMP_COMMUNITY=default",
		"SNMP_VERSION=2c",
		"EMAIL_FROM=fz@example",
	}
	got := cfg.Tasks[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.FrequencySeconds != want.FrequencySeconds {
		t.Errorf("got %d, want %d", got.FrequencySeconds, want.FrequencySeconds)
	}

	if got.Frequency() != time.Duration(want.FrequencySeconds)*time.Second {
		t.Errorf("got %d, want %d", got.Frequency(), time.Duration(want.FrequencySeconds)*time.Second)
	}

	if fmt.Sprintf("%v", got.Environment()) != fmt.Sprintf("%s", wantEnvironment) {
		t.Errorf("got %v, want %v", got.Environment(), wantEnvironment)
	}
}

func TestConfigNotifierLogger(t *testing.T) {
	ReadConfig(
		fmt.Sprintf("%s/example_config.toml", workDir),
		fmt.Sprintf("%s/libexec", workDir),
		DEFAULT_LOG_FILE,
		DEFAULT_LOG_LEVEL,
	)

	want := Notifier{
		Name:           "logger",
		Command:        "notifier/null",
		TimeoutSeconds: 1,
		GateSetStrings: [][]string{
			[]string{"is_not_unknown", "to_failed", "defer"},
			[]string{"is_not_unknown", "is_failed", "renotify"},
		},
		Envs: []string{},
	}
	got := cfg.Notifiers[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
	}

	if fmt.Sprintf("%v", got.GateSetStrings) != fmt.Sprintf("%s", want.GateSetStrings) {
		t.Errorf("got %v, want %v", got.GateSetStrings, want.GateSetStrings)
	}

	if fmt.Sprintf("%v", got.Envs) != fmt.Sprintf("%s", want.Envs) {
		t.Errorf("got %v, want %v", got.Envs, want.Envs)
	}
}

func TestConfigNotifierErrorEmailer(t *testing.T) {
	ReadConfig(
		fmt.Sprintf("%s/example_config.toml", workDir),
		fmt.Sprintf("%s/libexec", workDir),
		DEFAULT_LOG_FILE,
		DEFAULT_LOG_LEVEL,
	)

	want := Notifier{
		Name:           "error_emailer",
		Command:        "notifier/email",
		TimeoutSeconds: 3,
		GateSetStrings: [][]string{},
		Envs: []string{
			"EMAIL_ADDRESS=jon@altos.au",
			"EMAIL_FROM=fz@altos.au",
			"EMAIL_SUBJECT='fz experienced a critical error'",
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}

	wantEnvironment := []string{
		"EMAIL_ADDRESS=root@example",
		"EMAIL_FROM=fz_errors@example",
		"EMAIL_SUBJECT='fz experienced a critical error'",
		"SNMP_COMMUNITY=default",
		"SNMP_VERSION=2c",
	}

	got := cfg.Notifiers[1]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
	}

	if fmt.Sprintf("%v", got.GateSetStrings) != fmt.Sprintf("%s", want.GateSetStrings) {
		t.Errorf("got %v, want %v", got.GateSetStrings, want.GateSetStrings)
	}

	if fmt.Sprintf("%v", got.Environment()) != fmt.Sprintf("%s", wantEnvironment) {
		t.Errorf("got %v, want %v", got.Environment(), wantEnvironment)
	}
}

func TestConfigGateToFailed(t *testing.T) {
	ReadConfig(
		fmt.Sprintf("%s/example_config.toml", workDir),
		fmt.Sprintf("%s/libexec", workDir),
		DEFAULT_LOG_FILE,
		DEFAULT_LOG_LEVEL,
	)

	want := Gate{
		Name:    "to_failed",
		Command: "gate/to_state",
		Args:    []string{"fail"},
		Envs: []string{
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}

	wantEnvironment := []string{
		"GATE_NAME=to_failed",
		"GATE_TIMEOUT=1",
		"SNMP_COMMUNITY=default",
		"SNMP_VERSION=2c",
		"EMAIL_FROM=fz@example",
	}

	got := cfg.Gates[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if fmt.Sprintf("%v", got.Args) != fmt.Sprintf("%s", want.Args) {
		t.Errorf("got %v, want %v", got.Args, want.Args)
	}

	if fmt.Sprintf("%v", got.environment(nil)) != fmt.Sprintf("%s", wantEnvironment) {
		t.Errorf("got %v, want %v", got.environment(nil), wantEnvironment)
	}
}

///////////////////////////////////////////////////////////////////////////////
// VALIDATOR CHECKS

func TestConfigValidateNotifiersExistDefault(t *testing.T) {
	cfg = Config{
		Defaults: ConfigDefaults{
			NotifierNames: []string{"dont_exist"},
		},
	}

	want := ErrNotExist
	got := errors.Unwrap(cfg.validateNotifiersExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateNotifiersExistForTask(t *testing.T) {
	cfg = Config{
		Tasks: []Task{
			Task{NotifierNames: []string{"dont_exist"}},
		},
	}

	want := ErrNotExist
	got := errors.Unwrap(cfg.validateNotifiersExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateGatesExistForNotifier(t *testing.T) {
	cfg = Config{
		Notifiers: []Notifier{
			Notifier{GateSetStrings: [][]string{[]string{"dont_exist"}}},
		},
	}

	want := ErrNotExist
	got := errors.Unwrap(cfg.validateGatesExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateCommandsExistsForTask(t *testing.T) {
	cfg = Config{
		Tasks: []Task{
			Task{Command: "dont_exist"},
		},
	}

	want := ErrCommandNotExist
	got := errors.Unwrap(cfg.validateCommandsExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateCommandsExistsForNotifier(t *testing.T) {
	cfg = Config{
		Notifiers: []Notifier{
			Notifier{Command: "dont_exist"},
		},
	}

	want := ErrCommandNotExist
	got := errors.Unwrap(cfg.validateCommandsExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateCommandsExistsForGate(t *testing.T) {
	cfg = Config{
		Gates: []Gate{
			Gate{Command: "dont_exist"},
		},
	}

	want := ErrCommandNotExist
	got := errors.Unwrap(cfg.validateCommandsExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateNameForNotifier(t *testing.T) {
	cfg = Config{
		Notifiers: []Notifier{
			Notifier{Name: "with spaces"},
		},
	}

	want := ErrInvalidName
	got := errors.Unwrap(cfg.validateName())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateFrequencySecondsDefault(t *testing.T) {
	cfg = Config{
		Defaults: ConfigDefaults{},
	}

	want := ErrLessThan1
	got := errors.Unwrap(cfg.validateFrequencySeconds())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateTimeoutSecondsTask(t *testing.T) {
	cfg = Config{
		Defaults: ConfigDefaults{
			FrequencySeconds:      5,
			RetryFrequencySeconds: 5,
			TimeoutSeconds:        10,
		},
	}

	want := ErrTimeoutSlowerThanRetry
	got := errors.Unwrap(cfg.validateTimeoutSeconds())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidatePriorityHigh(t *testing.T) {
	cfg = Config{
		Defaults: ConfigDefaults{Priority: 100},
	}

	want := ErrGreaterThan99
	got := errors.Unwrap(cfg.validatePriority())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidatePriorityLow(t *testing.T) {
	cfg = Config{
		Defaults: ConfigDefaults{Priority: 0},
		Tasks: []Task{
			Task{},
		},
	}

	want := ErrLessThan1
	got := errors.Unwrap(cfg.validatePriority())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
