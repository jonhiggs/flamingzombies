package fz

import (
	"io"
	"os"

	"github.com/pelletier/go-toml"
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
			config.Tasks[i].Timeout = config.Defaults.Timeout
		}

		// prepare to receive state signals
		States = append(States, State{t.Hash(), t.Retries, 0b10})
	}

	return config
}
