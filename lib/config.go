package fz

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300

type Defaults struct {
	Notifiers        []string `toml:"notifiers"`
	Retries          int      `toml:"retries"`
	FrequencySeconds int      `toml:"frequency_seconds"`
	TimeoutSeconds   int      `toml:"timeout_seconds"` // better to put the timeout into the commmand
}

type Config struct {
	Defaults  Defaults
	LogLevel  string     `toml:"log_level"`
	Notifiers []Notifier `toml:"notifiers"`
	Tasks     []Task     `toml:"task"`
}

func ReadConfig() Config {
	file, err := os.Open("config.toml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var config Config

	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}

	// set the default defaults
	if config.Defaults.Retries == 0 {
		config.Defaults.Retries = DEFAULT_RETRIES
	}

	if config.Defaults.TimeoutSeconds == 0 {
		config.Defaults.TimeoutSeconds = DEFAULT_TIMEOUT_SECONDS
	}

	if config.Defaults.FrequencySeconds == 0 {
		config.Defaults.FrequencySeconds = DEFAULT_FREQUENCY_SECONDS
	}

	for i, t := range config.Tasks {
		// fill in the defaults
		if len(t.Notifiers) == 0 && len(config.Defaults.Notifiers) != 0 {
			t.NotifierStr = config.Defaults.Notifiers
		}

		if t.Retries == 0 {
			config.Tasks[i].Retries = config.Defaults.Retries
		}

		if t.TimeoutSeconds == 0 {
			config.Tasks[i].TimeoutSeconds = config.Defaults.TimeoutSeconds
		}

		if t.FrequencySeconds == 0 {
			config.Tasks[i].FrequencySeconds = config.Defaults.FrequencySeconds
		}

		// create the duration from the *_seconds settings
		config.Tasks[i].Timeout = time.Duration(config.Tasks[i].TimeoutSeconds) * time.Second

		// construct the Task.Notifiers
		for _, ns := range config.Tasks[i].NotifierStr {
			config.Tasks[i].Notifiers = append(config.Tasks[i].Notifiers, config.FindNotifier(ns))
		}

		// start the history in an unknown state
		config.Tasks[i].history = 0b10

		// validate the inputs
		if config.Tasks[i].Retries > 32 {
			log.WithFields(log.Fields{
				"file":      "lib/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal("cannot retry more than 32 times")
		}

		if config.Tasks[i].FrequencySeconds < 1 {
			log.WithFields(log.Fields{
				"file":      "lib/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal("frequency_seconds must be greater than 1")
		}

		if config.Tasks[i].TimeoutSeconds > config.Tasks[i].FrequencySeconds {
			log.WithFields(log.Fields{
				"file":      "lib/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal(fmt.Sprintf("frequency_seconds (%d) must be shorter than the timeout_seconds (%d)", config.Tasks[i].FrequencySeconds, config.Tasks[i].TimeoutSeconds))
		}
	}

	return config
}

func (c Config) FindNotifier(s string) *Notifier {
	for i, _ := range c.Notifiers {
		if s == c.Notifiers[i].Name {
			return &c.Notifiers[i]
		}
	}

	log.WithFields(log.Fields{
		"file":          "lib/config.go",
		"notifier_name": s,
	}).Fatal("unknown notifier")
	return nil
}
