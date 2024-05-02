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
	GateSets       [][]string `toml:"gates"`
	TimeoutSeconds int        `toml:"timeout"`
}

func (n Notifier) timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notifier) gates() [][]*Gate {
	gateSets := make([][]*Gate, len(n.GateSets))
	for i := 0; i < len(n.GateSets); i++ {
		gateSets[i] = make([]*Gate, 30)
	}

	for gsi, gs := range n.GateSets {
		for gi, gn := range gs {
			g, _ := GateByName(gn)
			gateSets[gsi][gi] = g
		}
	}

	return gateSets
}

func (n Notifier) validate() error {
	if _, err := os.Stat(n.Command); os.IsNotExist(err) {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, n.Command)); os.IsNotExist(err) {
			return fmt.Errorf("notifier command not found")
		}
	}

	// TODO: make sure that no GateSet has more than 30 elements.

	return nil
}
