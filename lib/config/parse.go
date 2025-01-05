package config

import (
	"github.com/BurntSushi/toml"
)

// return a configuration from the []byte returned from PreProcess.
func Parse() (Config, error) {

	return Config{}, nil
}

// return the default values from the []byte returned from PreProcess.
func parseDefault(b []byte) (defaultConfig, error) {
	var cfg Config

	err := toml.Unmarshal(b, &cfg)
	if err != nil {
		return defaultConfig{}, err
	}

	return cfg.Def, nil
}
