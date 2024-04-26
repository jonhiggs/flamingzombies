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
const DEFAULT_PRIORITY = 5
const DEFAULT_UNKNOWN_EXIT_CODE = 3 // straight from Nagios

type Defaults struct {
	FrequencySeconds      int      `toml:"frequency_seconds"`
	NotifierNames         []string `toml:"notifiers"`
	Retries               int      `toml:"retries"`
	RetryFrequencySeconds int      `toml:"retry_frequency_seconds"`
	TimeoutSeconds        int      `toml:"timeout_seconds"` // better to put the timeout into the commmand
	Priority              int      `toml:"priority"`
	UnknownExitCode       int      `toml:"unknown_exit_code`
}

type Config struct {
	Defaults      Defaults
	LogLevel      string     `toml:"log_level"`
	LogFile       string     `toml:"log_file"`
	Notifiers     []Notifier `toml:"notifier"`
	Tasks         []Task     `toml:"task"`
	Gates         []Gate     `toml:"gate"`
	ListenAddress string     `toml:"listen_address"`
	Directory     string     `toml:"directory"`
}

var config Config

func ReadConfig() Config {
	file, err := os.Open(os.Getenv("FZ_CONFIG_FILE"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening configuration file %s\n", os.Getenv("FZ_CONFIG_FILE"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration file %s\n", os.Getenv("FZ_CONFIG_FILE"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration file %s\n", os.Getenv("FZ_CONFIG_FILE"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if config.LogFile == "" {
		config.LogFile = os.Getenv("FZ_LOG_FILE")
	}

	if config.LogLevel == "" {
		config.LogLevel = os.Getenv("FZ_LOG_LEVEL")
	}

	if config.Directory == "" {
		config.LogLevel = os.Getenv("FZ_DIRECTORY")
	}

	if config.ListenAddress == "" {
		config.LogLevel = os.Getenv("FZ_LISTEN")
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
				config.Tasks[i].RetryFrequencySeconds = config.Tasks[i].TimeoutSeconds
			} else {
				config.Tasks[i].RetryFrequencySeconds = config.Defaults.RetryFrequencySeconds
			}
		}

		if t.Priority == 0 {
			if config.Defaults.Priority == 0 {
				config.Tasks[i].Priority = DEFAULT_PRIORITY
			} else {
				config.Tasks[i].Priority = config.Defaults.Priority
			}
		}

		if t.UnknownExitCode == 0 {
			if config.Defaults.UnknownExitCode == 0 {
				config.Tasks[i].UnknownExitCode = DEFAULT_UNKNOWN_EXIT_CODE
			} else {
				config.Tasks[i].Priority = config.Defaults.UnknownExitCode
			}
		}

		if len(t.ErrorBody) == 0 {
			config.Tasks[i].ErrorBody = "The task has entered an error state"
		}
		if len(t.RecoverBody) == 0 {
			config.Tasks[i].RecoverBody = "The task has recovered from an error state"
		}

		if len(t.NotifierNames) == 0 {
			config.Tasks[i].NotifierNames = config.Defaults.NotifierNames
		}

		// start the history in an unknown state
		config.Tasks[i].History = 0b10

		// validate the inputs
		if config.Tasks[i].Retries > 32 {
			log.WithFields(log.Fields{
				"file":      "lib/fz/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal("cannot retry more than 32 times")
		}

		if config.Tasks[i].FrequencySeconds < 1 {
			log.WithFields(log.Fields{
				"file":      "lib/fz/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal("frequency_seconds must be greater than 1")
		}

		if config.Tasks[i].TimeoutSeconds > config.Tasks[i].FrequencySeconds {
			log.WithFields(log.Fields{
				"file":      "lib/fz/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal(fmt.Sprintf("frequency_seconds (%d) must be shorter than the timeout_seconds (%d)", config.Tasks[i].FrequencySeconds, config.Tasks[i].TimeoutSeconds))
		}

		if config.Tasks[i].TimeoutSeconds > config.Tasks[i].RetryFrequencySeconds {
			log.WithFields(log.Fields{
				"file":      "lib/fz/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal(fmt.Sprintf("retry_frequency_seconds (%d) must be shorter than the timeout_seconds (%d)", config.Tasks[i].RetryFrequencySeconds, config.Tasks[i].TimeoutSeconds))
		}

		if config.Tasks[i].Priority < 0 || config.Tasks[i].Priority > 100 {
			log.WithFields(log.Fields{
				"file":      "lib/fz/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal("priority must be between 1 and 100")
		}

		// hit the notifiers() method to check that all specified notifiers exist
		config.Tasks[i].notifiers()

		if err = t.validate(); err != nil {
			log.WithFields(log.Fields{
				"file":      "lib/fz/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal(err)
		}
	}

	return config
}

func (c Config) Listen() bool {
	return c.ListenAddress != ""
}
