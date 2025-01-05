package config

var Directory, LogFile, LogLevel string

type Config struct {
	//Gates     []Gate         `toml:"gate"`
	//Notifiers []Notifier     `toml:"notifier"`
	//Tasks     []Task         `toml:"task"`

	Directory string        `toml:"directory"`
	LogFile   string        `toml:"log_file"`
	LogLevel  string        `toml:"log_level"`
	Def       defaultConfig `toml:"defaults"`
}
