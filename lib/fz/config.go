package fz

import (
	"fmt"
	"io"
	"os"

	"github.com/pelletier/go-toml"
)

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300
const DEFAULT_PRIORITY = 5

type Defaults struct {
	FrequencySeconds      int      `toml:"frequency"`
	NotifierNames         []string `toml:"notifiers"`
	Retries               int      `toml:"retries"`
	RetryFrequencySeconds int      `toml:"retry_frequency"`
	TimeoutSeconds        int      `toml:"timeout"` // better to put the timeout into the commmand
	Priority              int      `toml:"priority"`
}

type Config struct {
	Defaults      Defaults
	LogLevel      string     `toml:"log_level"`
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

	if config.LogLevel == "" {
		config.LogLevel = os.Getenv("FZ_LOG_LEVEL")
	}

	if config.Directory == "" {
		config.Directory = os.Getenv("FZ_DIRECTORY")
	}

	if config.ListenAddress == "" {
		config.ListenAddress = os.Getenv("FZ_LISTEN")
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
			panic(fmt.Sprintf("task '%s' cannot retry more than 32 times", t.Name))
		}

		if config.Tasks[i].FrequencySeconds < 1 {
			panic(fmt.Sprintf("task '%s' must have a frequency greater than 0", t.Name))
		}

		if config.Tasks[i].TimeoutSeconds > config.Tasks[i].FrequencySeconds {
			panic(fmt.Sprintf("task '%s' must have its timeout shorter than its frequency", t.Name))
		}

		if config.Tasks[i].TimeoutSeconds > config.Tasks[i].RetryFrequencySeconds {
			panic(fmt.Sprintf("task '%s' must have its timeout shorter than its retry_frequency", t.Name))
		}

		if config.Tasks[i].Priority < 0 || config.Tasks[i].Priority > 100 {
			panic(fmt.Sprintf("task '%s' must a priority between 1 and 100", t.Name))
		}

		// hit the notifiers() method to check that all specified notifiers exist
		config.Tasks[i].notifiers()

		if err = t.validate(); err != nil {
			panic(fmt.Sprintf("task '%s': %s", t.Name, err))
		}
	}

	for _, n := range config.Notifiers {
		if err = n.validate(); err != nil {
			panic(fmt.Sprintf("notifier '%s': %s", n.Name, err))
		}

	}

	for _, g := range config.Gates {
		if err = g.validate(); err != nil {
			panic(fmt.Sprintf("gate '%s': %s", g.Name, err))
		}
	}

	return config
}

func (c Config) Listen() bool {
	return c.ListenAddress != ""
}
