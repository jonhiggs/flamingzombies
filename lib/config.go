package fz

import (
	"fmt"
	"io"
	"os"

	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300

type Defaults struct {
	FrequencySeconds      int      `toml:"frequency_seconds"`
	NotifierNames         []string `toml:"notifiers"`
	Retries               int      `toml:"retries"`
	RetryFrequencySeconds int      `toml:"retry_frequency_seconds"`
	TimeoutSeconds        int      `toml:"timeout_seconds"` // better to put the timeout into the commmand
}

type Config struct {
	Defaults  Defaults
	LogLevel  string     `toml:"log_level"`
	Notifiers []Notifier `toml:"notifier"`
	Tasks     []Task     `toml:"task"`
}

var config Config

func ReadConfig() Config {
	file, err := os.Open("config.toml")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		panic(err)
	}

	for i, t := range config.Tasks {
		if t.Retries == 0 {
			if config.Defaults.Retries == 0 {
				config.Tasks[i].Retries = DEFAULT_RETRIES
			} else {
				config.Tasks[i].Retries = config.Defaults.Retries
			}
		}

		if t.TimeoutSeconds == 0 {
			if config.Defaults.TimeoutSeconds == 0 {
				config.Tasks[i].TimeoutSeconds = DEFAULT_TIMEOUT_SECONDS
			} else {
				config.Tasks[i].TimeoutSeconds = config.Defaults.TimeoutSeconds
			}
		}

		if t.FrequencySeconds == 0 {
			if config.Defaults.FrequencySeconds == 0 {
				config.Tasks[i].FrequencySeconds = DEFAULT_FREQUENCY_SECONDS
			} else {
				config.Tasks[i].FrequencySeconds = config.Defaults.FrequencySeconds
			}
		}

		if t.RetryFrequencySeconds == 0 {
			if config.Defaults.RetryFrequencySeconds == 0 {
				config.Tasks[i].RetryFrequencySeconds = config.Tasks[i].FrequencySeconds
			} else {
				config.Tasks[i].RetryFrequencySeconds = config.Defaults.RetryFrequencySeconds
			}
		}

		if len(t.NotifierNames) == 0 {
			config.Tasks[i].NotifierNames = config.Defaults.NotifierNames
		}

		// start the history in an unknown state
		config.Tasks[i].history = 0b10
		config.Tasks[i].lastState = STATE_UNKNOWN
		config.Tasks[i].stateChanged = false

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

		// hit the notifiers() method to check that all specified notifiers exist
		config.Tasks[i].notifiers()
	}

	return config
}
