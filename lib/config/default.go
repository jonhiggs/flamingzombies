package config

import "time"

var def defaultConfig

type defaultConfig struct {
	Envs                  []string `toml:"envs"`
	ErrorNotifierNames    []string `toml:"error_notifiers"`
	FrequencySeconds      int      `toml:"frequency"`
	NotifierNames         []string `toml:"notifiers"`
	Priority              int      `toml:"priority"`
	Retries               int      `toml:"retries"`
	RetryFrequencySeconds int      `toml:"retry_frequency"`
	TimeoutSeconds        int      `toml:"timeout"` // better to put the timeout into the command
}

func DefaultTimemout() time.Duration {
	return time.Duration(def.TimeoutSeconds) * time.Second
}
