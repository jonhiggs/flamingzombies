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
	config = ReadConfig(fmt.Sprintf("%s/example_config.toml", workDir))
	config.Directory = fmt.Sprintf("%s/libexec", workDir)

	wantLogFile := "-"
	wantLogLevel := "info"
	got := config

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

	if len(got.Notifiers) != 2 {
		t.Errorf("got %d, want %d", len(got.Notifiers), 2)
	}

	if len(got.Gates) != 5 {
		t.Errorf("got %d, want %d", len(got.Gates), 5)
	}

	if err := config.Validate(); err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}
}

func TestConfigDefaults(t *testing.T) {
	config = ReadConfig(fmt.Sprintf("%s/example_config.toml", workDir))
	config.Directory = fmt.Sprintf("%s/libexec", workDir)
	want := ConfigDefaults{
		Retries:            5,
		TimeoutSeconds:     1,
		NotifierNames:      []string{"logger"},
		ErrorNotifierNames: []string{"error_emailer"},
		Priority:           3,
		FrequencySeconds:   300,
		Envs: []string{
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}
	got := config.Defaults

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
	config = ReadConfig(fmt.Sprintf("%s/example_config.toml", workDir))
	config.Directory = fmt.Sprintf("%s/libexec", workDir)
	want := Task{
		Name:             "flappy",
		Command:          "task/flappy",
		FrequencySeconds: 1,
		ErrorBody:        "flappy has entered an error state\n",
		RecoverBody:      "flappy has recovered\n",
		Envs: []string{
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}
	got := config.Tasks[0]

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

	if got.ErrorBody != want.ErrorBody {
		t.Errorf("got %s, want %s", got.ErrorBody, want.ErrorBody)
	}

	if got.RecoverBody != want.RecoverBody {
		t.Errorf("got %s, want %s", got.RecoverBody, want.RecoverBody)
	}

	if fmt.Sprintf("%v", got.Envs) != fmt.Sprintf("%s", want.Envs) {
		t.Errorf("got %v, want %v", got.Envs, want.Envs)
	}
}

func TestConfigNotifierLogger(t *testing.T) {
	config = ReadConfig(fmt.Sprintf("%s/example_config.toml", workDir))
	config.Directory = fmt.Sprintf("%s/libexec", workDir)
	want := Notifier{
		Name:           "logger",
		Command:        "notifier/null",
		TimeoutSeconds: 5,
		GateSets: [][]string{
			[]string{"to_failed", "defer"},
			[]string{"is_failed", "renotify"},
		},
		Envs: []string{
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}
	got := config.Notifiers[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
	}

	if fmt.Sprintf("%v", got.GateSets) != fmt.Sprintf("%s", want.GateSets) {
		t.Errorf("got %v, want %v", got.GateSets, want.GateSets)
	}

	if fmt.Sprintf("%v", got.Envs) != fmt.Sprintf("%s", want.Envs) {
		t.Errorf("got %v, want %v", got.Envs, want.Envs)
	}
}

func TestConfigNotifierErrorEmailer(t *testing.T) {
	config = ReadConfig(fmt.Sprintf("%s/example_config.toml", workDir))
	config.Directory = fmt.Sprintf("%s/libexec", workDir)
	want := Notifier{
		Name:           "error_emailer",
		Command:        "notifier/email",
		TimeoutSeconds: 3,
		GateSets:       [][]string{},
		Envs: []string{
			"EMAIL_ADDRESS=jon@altos.au",
			"EMAIL_FROM=fz@altos.au",
			"EMAIL_SUBJECT='fz experienced a critical error'",
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}
	got := config.Notifiers[1]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if got.TimeoutSeconds != want.TimeoutSeconds {
		t.Errorf("got %d, want %d", got.TimeoutSeconds, want.TimeoutSeconds)
	}

	if fmt.Sprintf("%v", got.GateSets) != fmt.Sprintf("%s", want.GateSets) {
		t.Errorf("got %v, want %v", got.GateSets, want.GateSets)
	}

	if fmt.Sprintf("%v", got.Envs) != fmt.Sprintf("%s", want.Envs) {
		t.Errorf("got %v, want %v", got.Envs, want.Envs)
	}
}

func TestConfigGateToFailed(t *testing.T) {
	config = ReadConfig(fmt.Sprintf("%s/example_config.toml", workDir))
	config.Directory = fmt.Sprintf("%s/libexec", workDir)
	want := Gate{
		Name:    "to_failed",
		Command: "gate/to_state",
		Args:    []string{"fail"},
		Envs: []string{
			"SNMP_COMMUNITY=default",
			"SNMP_VERSION=2c",
		},
	}
	got := config.Gates[0]

	if got.Name != want.Name {
		t.Errorf("got %s, want %s", got.Name, want.Name)
	}

	if got.Command != want.Command {
		t.Errorf("got %s, want %s", got.Command, want.Command)
	}

	if fmt.Sprintf("%v", got.Args) != fmt.Sprintf("%s", want.Args) {
		t.Errorf("got %v, want %v", got.Args, want.Args)
	}

	if fmt.Sprintf("%v", got.Envs) != fmt.Sprintf("%s", want.Envs) {
		t.Errorf("got %v, want %v", got.Envs, want.Envs)
	}
}

///////////////////////////////////////////////////////////////////////////////
// VALIDATOR CHECKS

func TestConfigValidateNotifiersExistDefault(t *testing.T) {
	config = Config{
		Defaults: ConfigDefaults{
			NotifierNames: []string{"dont_exist"},
		},
	}

	want := ErrNotExist
	got := errors.Unwrap(config.validateNotifiersExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateNotifiersExistForTask(t *testing.T) {
	config = Config{
		Tasks: []Task{
			Task{NotifierNames: []string{"dont_exist"}},
		},
	}

	want := ErrNotExist
	got := errors.Unwrap(config.validateNotifiersExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateGatesExistForNotifier(t *testing.T) {
	config = Config{
		Notifiers: []Notifier{
			Notifier{GateSets: [][]string{[]string{"dont_exist"}}},
		},
	}

	want := ErrNotExist
	got := errors.Unwrap(config.validateGatesExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateCommandsExistsForTask(t *testing.T) {
	config = Config{
		Tasks: []Task{
			Task{Command: "dont_exist"},
		},
	}

	want := ErrCommandNotExist
	got := errors.Unwrap(config.validateCommandsExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateCommandsExistsForNotifier(t *testing.T) {
	config = Config{
		Notifiers: []Notifier{
			Notifier{Command: "dont_exist"},
		},
	}

	want := ErrCommandNotExist
	got := errors.Unwrap(config.validateCommandsExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateCommandsExistsForGate(t *testing.T) {
	config = Config{
		Gates: []Gate{
			Gate{Command: "dont_exist"},
		},
	}

	want := ErrCommandNotExist
	got := errors.Unwrap(config.validateCommandsExist())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateNameForNotifier(t *testing.T) {
	config = Config{
		Notifiers: []Notifier{
			Notifier{Name: "with spaces"},
		},
	}

	want := ErrInvalidName
	got := errors.Unwrap(config.validateName())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateFrequencySecondsDefault(t *testing.T) {
	config = Config{
		Defaults: ConfigDefaults{},
	}

	want := ErrLessThan1
	got := errors.Unwrap(config.validateFrequencySeconds())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidateTimeoutSecondsTask(t *testing.T) {
	config = Config{
		Defaults: ConfigDefaults{
			FrequencySeconds:      5,
			RetryFrequencySeconds: 5,
			TimeoutSeconds:        10,
		},
	}

	want := ErrTimeoutSlowerThanRetry
	got := errors.Unwrap(config.validateTimeoutSeconds())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidatePriorityHigh(t *testing.T) {
	config = Config{
		Defaults: ConfigDefaults{Priority: 100},
	}

	want := ErrGreaterThan99
	got := errors.Unwrap(config.validatePriority())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestConfigValidatePriorityLow(t *testing.T) {
	config = Config{
		Defaults: ConfigDefaults{Priority: 0},
		Tasks: []Task{
			Task{},
		},
	}

	want := ErrLessThan1
	got := errors.Unwrap(config.validatePriority())
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
