package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/BurntSushi/toml"
)

var Directory = "/usr/lib/flamingzombies"
var LogFile = "-"
var LogLevel = "info"

// The default values to insert into the task, gate, and notifier resources
// when needed.
var def defaultConfig

// TODO(jh) 20250106: the resources
//var Tasks []Task
//var Notifiers []Notifier
//var Gates []Gate

type tomlConfig struct {
	//Gates     []Gate         `toml:"gate"`
	//Notifiers []Notifier     `toml:"notifier"`
	//Tasks     []Task         `toml:"task"`

	Directory string        `toml:"directory"`
	LogFile   string        `toml:"log_file"`
	LogLevel  string        `toml:"log_level"`
	Default   defaultConfig `toml:"defaults"`
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

	// EXAMPLE: https://github.com/BurntSushi/toml/issues/47
	//config := Host{
	//	Servers: []Server{
	//		{
	//			Url:  "http://google.com",
	//			Port: 80,
	//		},
	//	},
	//}
	//if _, err := toml.Decode(blob, &config); err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("%#v\n", config)

	return nil
}

// Extract the named toml blocks from a configuration file. This is needed so
// that the default values can be applied to anything unset during the
// Decoding. Otherwise it's not possible to differenciate between a
// user-declared 0 or a default value of 0. The former should be used, the
// latter should be replaced with the default value.
func extractTomlOjbects(n string, b []byte) [][]byte {
	startBlock := regexp.MustCompile(fmt.Sprintf(`^\[\[%s\]\]`, n))
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

func tomlTasks(b []byte) ([]Task, error) {
	defaultTask := Task{
		Args:                  def.Args,
		Command:               def.Command,
		Description:           def.Description,
		Envs:                  def.Envs,
		ErrorNotifierNames:    def.ErrorNotifierNames,
		FrequencySeconds:      def.FrequencySeconds,
		History:               0b010,
		HistoryMask:           0b111,
		LastFail:              time.Unix(0, 0),
		LastNotification:      time.Unix(0, 0),
		LastOk:                time.Unix(0, 0),
		LastRun:               time.Unix(0, 0),
		Name:                  def.Name,
		NotifierNames:         def.NotifierNames,
		Priority:              def.Priority,
		Retries:               def.Retries,
		RetryFrequencySeconds: def.RetryFrequencySeconds,
		TimeoutSeconds:        def.TimeoutSeconds,
	}

	var r []Task

	for _, blob := range extractTomlOjbects("task", b) {
		task := defaultTask
		_, err := toml.Decode(string(blob), &task)
		if err != nil {
			return []Task{}, err
		}

		r = append(r, task)

	}

	return r, nil
}
