package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

var Directory = "/usr/lib/flamingzombies"
var LogFile = "-"
var LogLevel = "info"

// The default values to insert into the task, gate, and notifier resources
// when needed.
var def defaultConfig

var Tasks []Task
var Gates []Gate

// TODO(jh) 20250106: the resources
//var Notifiers []Notifier
//var Gates []Gate

type tomlConfig struct {
	//Gates     []Gate         `toml:"gate"`
	//Notifiers []Notifier     `toml:"notifier"`
	//Tasks     []Task         `toml:"task"`

	Directory string        `toml:"directory"`
	LogFile   string        `toml:"log_file"`
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

// A Task is a command that is executed on a schedule. The struct contains the
// static configuration of the task which is read from the configuration file,
// and it's metadata and history which are generated over the course of the
// daemons lifecycle.
type Task struct {
	Name                  string   `toml:"name"`            // friendly name
	Description           string   `toml:"description"`     // description of the task
	Command               string   `toml:"command"`         // command
	Args                  []string `toml:"args"`            // command arguments
	FrequencySeconds      int      `toml:"frequency"`       // how often to run
	RetryFrequencySeconds int      `toml:"retry_frequency"` // how quickly to retry when state unknown
	TimeoutSeconds        int      `toml:"timeout"`         // how long an execution may run
	Retries               int      `toml:"retries"`         // number of retries before changing the state
	NotifierNames         []string `toml:"notifiers"`       // notifiers to trigger upon state change
	ErrorNotifierNames    []string `toml:"error_notifiers"` // notifiers to trigger upon state change
	Priority              int      `toml:"priority"`        // the priority of the notifications
	Envs                  []string `toml:"envs"`            // environment variables supplied to task

	// public, but not configurable
	History          uint32    // represented in binary. Successes are high
	HistoryMask      uint32    // the bits in the history with a recorded value. Needed to understand a history of 0
	LastFail         time.Time // the time of the last failed execution
	LastOk           time.Time // the time of the last successful execution
	LastNotification time.Time // the time of the last notification
	LastRun          time.Time // the time of the last execution
	TraceID          string    // the ID of the task execution to help with tracing
}

type Gate struct {
	Args    []string `toml:"args"`    // command arguments
	Command string   `toml:"command"` // command
	Envs    []string `toml:"envs"`    // environment variables
	Name    string   `toml:"name"`    // friendly name
}

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
		Directory = cfg.Directory
	}

	if cfg.LogFile != "" {
		LogFile = cfg.LogFile
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

	return nil
}

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
	defaultTask := Task{
		Args:                  d.Args,
		Command:               d.Command,
		Description:           d.Description,
		ErrorNotifierNames:    d.ErrorNotifierNames,
		FrequencySeconds:      d.FrequencySeconds,
		History:               0b010,
		HistoryMask:           0b111,
		LastFail:              time.Unix(0, 0),
		LastNotification:      time.Unix(0, 0),
		LastOk:                time.Unix(0, 0),
		LastRun:               time.Unix(0, 0),
		Name:                  d.Name,
		NotifierNames:         d.NotifierNames,
		Priority:              d.Priority,
		Retries:               d.Retries,
		RetryFrequencySeconds: d.RetryFrequencySeconds,
		TimeoutSeconds:        d.TimeoutSeconds,
	}

	var r []Task

	for _, blob := range extractTomlOjbects("task", b) {
		task := defaultTask
		_, err := toml.Decode(string(blob), &task)
		if err != nil {
			return []Task{}, err
		}

		// trim any trailing new lines
		task.Description = strings.TrimSuffix(task.Description, "\n")

		// the default merge of toml.Decode doesn't do what is needed.
		task.Envs = mergeEnvVars(task.Envs, d.Envs)

		r = append(r, task)

	}

	return r, nil
}

func gatesFromToml(b []byte, d defaultConfig) ([]Gate, error) {
	defaultGate := Gate{
		Args:    d.Args,
		Command: d.Command,
		Name:    d.Name,
	}

	var r []Gate

	for _, blob := range extractTomlOjbects("gate", b) {
		gate := defaultGate
		_, err := toml.Decode(string(blob), &gate)
		if err != nil {
			return []Gate{}, err
		}

		// the default merge of toml.Decode doesn't do what is needed.
		gate.Envs = mergeEnvVars(gate.Envs, d.Envs)

		r = append(r, gate)

	}

	return r, nil
}
