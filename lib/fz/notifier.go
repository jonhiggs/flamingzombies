package fz

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type Notifier struct {
	Name           string
	Command        string
	Args           []string
	GateNames      []string `toml:"gates"`
	TimeoutSeconds int      `toml:"timeout_seconds"`
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
			log.WithFields(log.Fields{
				"file":          "lib/fz/notifier.go",
				"notifier_name": n.Name,
			}).Fatal(fmt.Sprintf("unknown gate '%s'", gName))
		}
	}

	return gs
}
