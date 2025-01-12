package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

const VERSION = "v0.1.0"

var Directory string
var LogLevel = "info"

// The default values to insert into the task, gate, and notifier resources
// when needed.
var def defaultConfig

var Tasks []Task
var Gates []Gate
var Notifiers []Notifier

type tomlConfig struct {
	Directory string        `toml:"directory"`
	LogLevel  string        `toml:"log_level"`
	Default   defaultConfig `toml:"default"`
}

type defaultConfig struct {
	Args                  []string `toml:"args"`            // command arguments
	Command               string   `toml:"command"`         // command
	Description           string   `toml:"description"`     // description of the task
	Envs                  []string `toml:"envs"`            // environment variables supplied to task
	ErrorNotifierNames    []string `toml:"error_notifiers"` // notifiers to trigger upon state change
	FrequencySeconds      int      `toml:"frequency"`       // how often to run
	Name                  string   `toml:"name"`            // friendly name
	NotifierNames         []string `toml:"notifiers"`       // notifiers to trigger upon state change
	Priority              int      `toml:"priority"`        // the priority of the notifications
	Retries               int      `toml:"retries"`         // number of retries before changing the state
	RetryFrequencySeconds int      `toml:"retry_frequency"` // how quickly to retry when state unknown
	TimeoutSeconds        int      `toml:"timeout"`         // how long an execution may run
}

type State int8

const (
	STATE_UNKNOWN = State(iota)
	STATE_FAIL
	STATE_OK
)

func (s State) String() string {
	switch s {
	case STATE_FAIL:
		return "fail"
	case STATE_OK:
		return "ok"
	default:
		return "unknown"
	}
}

func init() {
	Directory, _ = os.Getwd()
}

// A Task is a command that is executed on a schedule. The struct contains the
// static configuration of the task which is read from the configuration file,
// and it's metadata and history which are generated over the course of the
// daemons lifecycle.

// populate the package variables from the content of the TOML
func Load(f *os.File) error {
	b, err := PreProcess(f, []*os.File{})
	if err != nil {
		return err
	}

	var cfg tomlConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return err
	}

	def = cfg.Default

	if cfg.Directory != "" {
		if err := SetDirectory(cfg.Directory); err != nil {
			return err
		}
	}

	if cfg.LogLevel != "" {
		LogLevel = cfg.LogLevel
	}

	Tasks, err = tasksFromToml(b, def)
	if err != nil {
		return err
	}

	Gates, err = gatesFromToml(b, def)
	if err != nil {
		return err
	}

	Notifiers, err = notifiersFromToml(b, def)
	if err != nil {
		return err
	}

	return nil
}

func SetDirectory(s string) error {
	Directory = s

	_, err := os.Stat(Directory)
	if err != nil {
		return fmt.Errorf("setting working directory: %w", errors.Unwrap(err))
	}

	return nil
}

/// PRIVATE ///////////////////////////////////////////////////////////////////

// Extract the named toml blocks from a configuration file. This is needed so
// that the default values can be applied to anything unset during the
// Decoding. Otherwise it's not possible to differenciate between a
// user-declared 0 or a default value of 0. The former should be used, the
// latter should be replaced with the default value.
func extractTomlOjbects(n string, b []byte) [][]byte {
	var startBlock *regexp.Regexp

	switch n {
	case "default":
		startBlock = regexp.MustCompile(fmt.Sprintf(`^\[%s\]`, n))
	case "task":
		startBlock = regexp.MustCompile(fmt.Sprintf(`^\[\[%s\]\]`, n))
	case "gate":
		startBlock = regexp.MustCompile(fmt.Sprintf(`^\[\[%s\]\]`, n))
	case "notifier":
		startBlock = regexp.MustCompile(fmt.Sprintf(`^\[\[%s\]\]`, n))
	}

	endBlock := regexp.MustCompile(`^\[`)
	inBlock := false

	scanner := bufio.NewScanner(bytes.NewReader(b))

	var objects [][]byte

	var o []byte

	for scanner.Scan() {
		l := scanner.Bytes()

		if endBlock.Match(l) {
			// if the object has data, flush it to the objects result
			if len(o) > 0 {
				objects = append(objects, o)
			}

			inBlock = false
			o = []byte{}
		}

		if startBlock.Match(l) {
			inBlock = true
			continue // remove the block heading
		}

		if inBlock {
			o = append(o, l...)
			o = append(o, byte('\n'))
		}
	}

	// flush any remaining data after scan has complete.
	if len(o) > 0 {
		objects = append(objects, o)
	}

	return objects
}

// convert the toml response from extractTomlOjbects into a defaultConfig.
func defaultConfigFromToml(b []byte) (defaultConfig, error) {
	var c defaultConfig
	if err := toml.Unmarshal(b, &c); err != nil {
		return defaultConfig{}, err
	}

	return c, nil
}

func tasksFromToml(b []byte, d defaultConfig) ([]Task, error) {
	var r []Task

	type taskData struct {
		Args                  []string `toml:"args"`            // command arguments
		Command               string   `toml:"command"`         // command
		Description           string   `toml:"description"`     // description of the task
		Envs                  []string `toml:"envs"`            // environment variables supplied to task
		ErrorNotifierNames    []string `toml:"error_notifiers"` // notifiers to trigger upon state change
		FrequencySeconds      int      `toml:"frequency"`       // how often to run
		Name                  string   `toml:"name"`            // friendly name
		NotifierNames         []string `toml:"notifiers"`       // notifiers to trigger upon state change
		Priority              int      `toml:"priority"`        // the priority of the notifications
		Retries               int      `toml:"retries"`         // number of retries before changing the state
		RetryFrequencySeconds int      `toml:"retry_frequency"` // how quickly to retry when state unknown
		TimeoutSeconds        int      `toml:"timeout"`         // how long an execution may run
	}

	for _, blob := range extractTomlOjbects("task", b) {
		task := taskData{
			Args:                  d.Args,
			Command:               d.Command,
			Description:           d.Description,
			ErrorNotifierNames:    d.ErrorNotifierNames,
			FrequencySeconds:      d.FrequencySeconds,
			NotifierNames:         d.NotifierNames,
			Priority:              d.Priority,
			Retries:               d.Retries,
			RetryFrequencySeconds: d.RetryFrequencySeconds,
			TimeoutSeconds:        d.TimeoutSeconds,
		}
		_, err := toml.Decode(string(blob), &task)
		if err != nil {
			return []Task{}, err
		}

		// trim any trailing new lines
		task.Description = strings.TrimSuffix(task.Description, "\n")

		// the default merge of toml.Decode doesn't do what is needed.
		task.Envs = mergeEnvVars(task.Envs, d.Envs)

		r = append(r, Task{
			args:                  task.Args,
			command:               task.Command,
			description:           task.Description,
			envs:                  mergeEnvVars(task.Envs, d.Envs),
			errorNotifierNames:    task.ErrorNotifierNames,
			frequencySeconds:      task.FrequencySeconds,
			name:                  task.Name,
			notifierNames:         task.NotifierNames,
			priority:              task.Priority,
			retries:               task.Retries,
			retryFrequencySeconds: task.RetryFrequencySeconds,
			timeoutSeconds:        task.TimeoutSeconds,
		})
	}

	return r, nil
}

// Extract and return gates from a toml configuration as []Gate
func gatesFromToml(b []byte, d defaultConfig) ([]Gate, error) {
	var r []Gate

	type gateData struct {
		Args    []string `toml:"args"`    // command arguments
		Command string   `toml:"command"` // command
		Envs    []string `toml:"envs"`    // environment variables
		Name    string   `toml:"name"`    // friendly name
	}

	for _, blob := range extractTomlOjbects("gate", b) {
		// set the default values
		gate := gateData{
			Args:    d.Args,
			Command: d.Command,
		}

		_, err := toml.Decode(string(blob), &gate)
		if err != nil {
			return []Gate{}, err
		}

		r = append(r, Gate{
			args:    gate.Args,
			command: gate.Command,
			envs:    mergeEnvVars(gate.Envs, d.Envs),
			name:    gate.Name,
		})
	}

	return r, nil
}

// Extract and return gates from a toml configuration as []Gate
func notifiersFromToml(b []byte, d defaultConfig) ([]Notifier, error) {
	var r []Notifier

	type notifierData struct {
		Args           []string   `toml:"args"`
		Command        string     `toml:"command"`
		Envs           []string   `toml:"envs"`
		GateSetStrings [][]string `toml:"gates"`
		Name           string     `toml:"name"`
		TimeoutSeconds int        `toml:"timeout"`
	}

	for _, blob := range extractTomlOjbects("notifier", b) {
		notifier := notifierData{
			Args:           d.Args,
			Command:        d.Command,
			TimeoutSeconds: d.TimeoutSeconds,
		}

		_, err := toml.Decode(string(blob), &notifier)
		if err != nil {
			return []Notifier{}, err
		}

		r = append(r, Notifier{
			args:           notifier.Args,
			command:        notifier.Command,
			envs:           mergeEnvVars(notifier.Envs, d.Envs),
			gateSetStrings: notifier.GateSetStrings,
			name:           notifier.Name,
			timeoutSeconds: notifier.TimeoutSeconds,
		})
	}

	return r, nil
}
