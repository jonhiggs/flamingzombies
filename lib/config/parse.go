package config

import (
	"github.com/BurntSushi/toml"
)

// return the default values from the []byte returned from PreProcess.
func getDefault(b []byte) (defaultConfig, error) {
	var cfg tomlConfig

	err := toml.Unmarshal(b, &cfg)
	if err != nil {
		return defaultConfig{}, err
	}

	return cfg.Default, nil
}
