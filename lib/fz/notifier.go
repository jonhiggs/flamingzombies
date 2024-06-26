package fz

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Notifier struct {
	Name           string
	Command        string
	Args           []string
	GateSets       [][]string `toml:"gates"`
	TimeoutSeconds int        `toml:"timeout"`

	gates [][]*Gate
}

func (n Notifier) timeout() time.Duration {
	return time.Duration(n.TimeoutSeconds) * time.Second
}

func (n Notifier) Gates() [][]*Gate {
	if len(n.gates) < 0 {
		return n.gates
	}

	n.gates = make([][]*Gate, len(n.GateSets))
	for i, gs := range n.GateSets {
		n.gates[i] = make([]*Gate, len(gs))
	}

	for gsi, gs := range n.GateSets {
		for gi, gn := range gs {
			g, _ := GateByName(gn)
			n.gates[gsi][gi] = g
		}
	}

	return n.gates
}

func (n Notifier) validate() error {
	if _, err := os.Stat(n.Command); os.IsNotExist(err) {
		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, n.Command)); os.IsNotExist(err) {
			return fmt.Errorf("notifier command not found")
		}
	}

	if strings.ContainsRune(n.Name, ' ') {
		return fmt.Errorf("name cannot contain spaces")
	}

	if strings.ContainsRune(n.Name, ',') {
		return fmt.Errorf("name cannot contain commas")
	}

	for i, gates := range n.GateSets {
		if len(gates) > 3 {
			return fmt.Errorf("gateset %d: cannot have more than 30 elements", i)
		}

		for _, g := range gates {
			_, err := GateByName(g)
			if err != nil {
				return fmt.Errorf("gateset %d: %s", i, err)
			}
		}
	}

	return nil
}

func NotifierByName(name string) *Notifier {
	for i, n := range config.Notifiers {
		if n.Name == name {
			return &config.Notifiers[i]
		}
	}

	return nil
}
