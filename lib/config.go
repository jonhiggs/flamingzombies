package fz

import (
	"io"
	"os"
	"time"

	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT = 5

type Defaults struct {
	Notifiers []string
	Retries   int
	Timeout   int
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
	if config.Defaults.Timeout == 0 {
		config.Defaults.Timeout = DEFAULT_TIMEOUT
	}

	// fill in the defaults
	for i, t := range config.Tasks {
		if len(t.Notifiers) == 0 && len(config.Defaults.Notifiers) != 0 {
			t.Notifiers = config.Defaults.Notifiers
		}

		if t.Retries == 0 {
			config.Tasks[i].Retries = config.Defaults.Retries
		}

		if t.Timeout == 0 {
			config.Tasks[i].Timeout = time.Duration(config.Defaults.Timeout) * time.Second
		}

		// start the history in an unknown state
		config.Tasks[i].History = 0b10

		// validate the inputs
		if config.Tasks[i].Retries > 32 {
			log.WithFields(log.Fields{
				"file":      "lib/config.go",
				"task_name": t.Name,
				"task_hash": t.Hash(),
			}).Fatal("cannot retry more than 32 times")
		}
	}

	return config
}
