package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

var Directory = "/usr/lib/flamingzombies"
var LogFile = "-"
var LogLevel = "info"

// The default values to insert into the task, gate, and notifier resources
// when needed.
var def defaultConfig

// TODO(jh) 20250106: the resources
//var Tasks []Task
//var Notifiers []Notifier
//var Gates []Gate

type tomlConfig struct {
	//Gates     []Gate         `toml:"gate"`
	//Notifiers []Notifier     `toml:"notifier"`
	//Tasks     []Task         `toml:"task"`

	Directory string        `toml:"directory"`
	LogFile   string        `toml:"log_file"`
	LogLevel  string        `toml:"log_level"`
	Default   defaultConfig `toml:"defaults"`
}

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

// populate the package variables from the content of the TOML
func Load(f *os.File) error {
	b, err := PreProcess(f, []*os.File{})
	if err != nil {
		return err
	}

	var cfg tomlConfig
	if err := toml.Unmarshal(b, &cfg); err != nil {
		return err
	}

	def = cfg.Default

	if cfg.Directory != "" {
		Directory = cfg.Directory
	}

	if cfg.LogFile != "" {
		LogFile = cfg.LogFile
	}

	if cfg.LogLevel != "" {
		LogLevel = cfg.LogLevel
	}

	return nil
}

// Extract the named toml blocks from a configuration file. This is needed so
// that the default values can be applied to anything unset during the
// Decoding. Otherwise it's not possible to differenciate between a
// user-declared 0 or a default value of 0. The former should be used, the
// latter should be replaced with the default value.
func extractTomlOjbects(n string, b []byte) [][]byte {

	// EXAMPLE: https://github.com/BurntSushi/toml/issues/47
	//config := Host{
	//	Servers: []Server{
	//		{
	//			Url:  "http://google.com",
	//			Port: 80,
	//		},
	//	},
	//}
	//if _, err := toml.Decode(blob, &config); err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Printf("%#v\n", config)

}
