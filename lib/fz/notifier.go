package fz

import (
	"fmt"
	"os"
	"time"
)

type Notifier struct {
	Name           string
	Command        string
	Args           []string
	GateNames      []string `toml:"gates"`
	TimeoutSeconds int      `toml:"timeout"`
}

func (n Notifier) timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notifier) gates() []*Gate {
	var gs []*Gate
	for _, gName := range n.GateNames {
		found := false
		for i, _ := range config.Gates {
			if gName == config.Gates[i].Name {
				gs = append(gs, &config.Gates[i])
				found = true
			}
		}

		if !found {
			panic(fmt.Sprintf("unknown gate '%s'", gName))
		}
	}

	return gs
}

func (n Notifier) validate() error {
	if _, err := os.Stat(n.Command); os.IsNotExist(err) {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, n.Command)); os.IsNotExist(err) {
			return fmt.Errorf("notifier command not found")
		}
	}

	return nil
}
