package fz

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/pelletier/go-toml"
)

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300
const DEFAULT_PRIORITY = 5

var DAEMON_START_TIME = time.Now()

var StatsdClient statsd.Statter = (*statsd.Client)(nil)

var Hostname string

var config Config

func ReadConfig() Config {
	Hostname, _ = os.Hostname()

	file, err := os.Open(os.Getenv("FZ_CONFIG_FILE"))
	if err != nil {
		Fatal(fmt.Sprintf("Error opening the configuration file %s\n", os.Getenv("FZ_CONFIG_FILE")), fmt.Sprint(err))
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		Fatal(fmt.Sprintf("Error reading the configuration file %s\n", os.Getenv("FZ_CONFIG_FILE")), fmt.Sprint(err))
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		Fatal(fmt.Sprintf("Error parsing the configuration file %s\n", os.Getenv("FZ_CONFIG_FILE")), fmt.Sprint(err))
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

	if config.StatsdHost == "" {
		config.StatsdHost = os.Getenv("FZ_STATSD_HOST")
	}

	if config.StatsdPrefix == "" {
		config.StatsdPrefix = os.Getenv("FZ_STATSD_PREFIX")
	}

	if config.StatsdHost != "" {
		StatsdClient, err = statsd.NewClient(config.StatsdHost, config.StatsdPrefix)
		if err != nil {
			panic(err)
		}
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

		// hit the notifiers() method to check that all specified notifiers exist
		config.Tasks[i].notifiers()

		if err = config.Tasks[i].validate(); err != nil {
			Fatal(fmt.Sprintf("task '%s': %s", t.Name, err))
		}
	}

	for _, n := range config.Notifiers {
		if err = n.validate(); err != nil {
			Fatal(fmt.Sprintf("notifier '%s': %s", n.Name, err))
		}
	}

	for _, g := range config.Gates {
		if err = g.validate(); err != nil {
			Fatal(fmt.Sprintf("gate '%s': %s", g.Name, err))
		}
	}

	return config
}

func (c Config) Listen() bool {
	return c.ListenAddress != ""
}

func (c Config) ErrorNotification() {
	//for _, n := range c.ErrorNotifiers {
	//	NotifyCh <- Notification{n, t}
	//}

}
