package config

var Directory, LogFile, LogLevel string

var Default defaultConfig

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
