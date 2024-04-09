package fz

import (
	"io"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Notifiers []Notifier `toml:"notifier"`
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

	return config
}
