package config

type Gate struct {
	args    []string // command arguments
	command string   // command
	envs    []string // environment variables
	name    string   // friendly name
}

func (g Gate) Args() []string {
	return g.args
}

func (g Gate) Command() string {
	return g.command
}

func (g Gate) Environment() []string {
	return g.envs
}

func (g Gate) Name() string {
	return g.name
}
