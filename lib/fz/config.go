package fz

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
)

const DEFAULT_RETRIES = 5
const DEFAULT_TIMEOUT_SECONDS = 5
const DEFAULT_FREQUENCY_SECONDS = 300
const DEFAULT_PRIORITY = 5

var DAEMON_START_TIME = time.Now()

var Hostname string

var config Config

func ReadConfig(f string) Config {
	Hostname, _ = os.Hostname()

	file, err := os.Open(f)
	if err != nil {
		Fatal(fmt.Sprintf("Error opening the configuration file '%s'\n", f), fmt.Sprint(err))
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err != nil {
		Fatal(fmt.Sprintf("Error reading the configuration file '%s'\n", f), fmt.Sprint(err))
	}

	err = toml.Unmarshal(b, &config)
	if err != nil {
		Fatal(fmt.Sprintf("Error parsing the configuration file '%s'\n", f), fmt.Sprint(err))
	}

	for i, t := range config.Tasks {
		if t.Retries == 0 {
			if config.Defaults.Retries == 0 {
				config.Tasks[i].Retries = DEFAULT_RETRIES
			} else {
				config.Tasks[i].Retries = config.Defaults.Retries
			}
		}

		if t.TimeoutSeconds == 0 {
			if config.Defaults.TimeoutSeconds == 0 {
				config.Tasks[i].TimeoutSeconds = DEFAULT_TIMEOUT_SECONDS
			} else {
				config.Tasks[i].TimeoutSeconds = config.Defaults.TimeoutSeconds
			}
		}

		if t.FrequencySeconds == 0 {
			if config.Defaults.FrequencySeconds == 0 {
				config.Tasks[i].FrequencySeconds = DEFAULT_FREQUENCY_SECONDS
			} else {
				config.Tasks[i].FrequencySeconds = config.Defaults.FrequencySeconds
			}
		}

		if t.RetryFrequencySeconds == 0 {
			if config.Defaults.RetryFrequencySeconds == 0 {
				config.Tasks[i].RetryFrequencySeconds = config.Tasks[i].TimeoutSeconds
			} else {
				config.Tasks[i].RetryFrequencySeconds = config.Defaults.RetryFrequencySeconds
			}
		}

		if t.Priority == 0 {
			if config.Defaults.Priority == 0 {
				config.Tasks[i].Priority = DEFAULT_PRIORITY
			} else {
				config.Tasks[i].Priority = config.Defaults.Priority
			}
		}

		for _, e := range config.Defaults.Envs {
			config.Tasks[i].Envs = append(config.Tasks[i].Envs, e)
		}

		if len(t.ErrorBody) == 0 {
			config.Tasks[i].ErrorBody = "The task has entered an error state"
		}
		if len(t.RecoverBody) == 0 {
			config.Tasks[i].RecoverBody = "The task has recovered from an error state"
		}

		if len(t.NotifierNames) == 0 {
			config.Tasks[i].NotifierNames = config.Defaults.NotifierNames
		}

		// start the history in an unknown state
		config.Tasks[i].History = 0b10

		// hit the notifiers() method to check that all specified notifiers exist
		config.Tasks[i].notifiers()
	}

	for i, _ := range config.Notifiers {
		for _, e := range config.Defaults.Envs {
			config.Notifiers[i].Envs = append(config.Notifiers[i].Envs, e)
		}
	}

	for i, _ := range config.Gates {
		for _, e := range config.Defaults.Envs {
			config.Gates[i].Envs = append(config.Gates[i].Envs, e)
		}
	}

	return config
}

func (c Config) Validate() error {
	//var err error

	//if err = config.validateTasks(); err != nil {
	//	return err
	//}

	//if err = config.validateNotifiers(); err != nil {
	//	return err
	//}

	//if err = config.validateGates(); err != nil {
	//	return err
	//}

	return nil
}

// Find and return a notifier by its name.
func (c Config) GetNotifierByName(name string) *Notifier {
	for i, n := range c.Notifiers {
		if n.Name == name {
			return &c.Notifiers[i]
		}
	}

	return nil
}

func (c Config) GetGateByName(name string) *Gate {
	for i, g := range c.Gates {
		if g.Name == name {
			return &c.Gates[i]
		}
	}

	return nil
}

// Return the GateSets attached to a notifier
func (c Config) GetNotifierGateSets(notifierName string) [][]*Gate {
	var r = [][]*Gate{}

	n := c.GetNotifierByName(notifierName)
	for igs, gateSet := range n.GateSets {
		for ig, gateName := range gateSet {
			g := config.GetGateByName(gateName)
			r[igs][ig] = g
		}
	}

	return r
}

func (c Config) validateTasks() error {
	for _, o := range config.Tasks {
		if err := o.validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) validateNotifiers() error {
	//for _, o := range config.Notifiers {
	//	if err := o.validate(); err != nil {
	//		return err
	//	}
	//}

	//func (n Notifier) validate() error {
	//	if _, err := os.Stat(n.Command); os.IsNotExist(err) {
	//		if _, err := os.Stat(fmt.Sprintf("%s/%s", config.Directory, n.Command)); os.IsNotExist(err) {
	//			return fmt.Errorf("notifier command not found")
	//		}
	//	}
	//
	//	if strings.ContainsRune(n.Name, ' ') {
	//		return fmt.Errorf("name cannot contain spaces")
	//	}
	//
	//	if strings.ContainsRune(n.Name, ',') {
	//		return fmt.Errorf("name cannot contain commas")
	//	}
	//
	//	for i, gates := range n.GateSets {
	//		if len(gates) > 30 {
	//			// TODO: why can't it have more than 30 elements?
	//			return fmt.Errorf("gateset %d: cannot have more than 30 elements", i)
	//		}
	//
	//		for _, g := range gates {
	//			_, err := GateByName(g)
	//			if err != nil {
	//				return fmt.Errorf("gateset %d: %s", i, err)
	//			}
	//		}
	//	}
	//
	//	return nil
	//}

	return nil
}

func (c Config) validateGates() error {
	return nil
}

func (c Config) ErrorNotification() {
	//for _, n := range c.ErrorNotifiers {
	//	NotifyCh <- Notification{n, t}
	//}

}

func (c Config) validateNotifiersExist() error {
	for _, n := range c.Defaults.NotifierNames {
		if c.GetNotifierByName(n) == nil {
			return fmt.Errorf("notifier %s: %w", n, ErrNotExist)
		}
	}

	for _, t := range c.Tasks {
		for _, n := range t.NotifierNames {
			if c.GetNotifierByName(n) == nil {
				return fmt.Errorf("notifier %s: %w", n, ErrNotExist)
			}
		}
	}

	return nil
}

func (c Config) validateGatesExist() error {
	for _, n := range c.Notifiers {
		for i, gs := range n.GateSets {
			for ii, g := range gs {
				if c.GetGateByName(g) == nil {
					return fmt.Errorf("gate [%d][%d]: %w", i, ii, ErrNotExist)
				}
			}
		}
	}

	return nil
}

func (c Config) validateCommandsExist() error {
	for i, t := range config.Tasks {
		if _, err := os.Stat(t.Command); os.IsNotExist(err) {
			return fmt.Errorf("task [%d]: command '%s': %w", i, t.Command, ErrCommandNotExist)
		}
	}

	for i, n := range config.Notifiers {
		if _, err := os.Stat(n.Command); os.IsNotExist(err) {
			return fmt.Errorf("notifier [%d]: command '%s': %w", i, n.Command, ErrCommandNotExist)
		}
	}

	for i, g := range config.Gates {
		if _, err := os.Stat(g.Command); os.IsNotExist(err) {
			return fmt.Errorf("gate [%d]: command '%s': %w", i, g.Command, ErrCommandNotExist)
		}
	}

	return nil
}

func (c Config) validateName() error {
	for i, t := range config.Tasks {
		if strings.ContainsRune(t.Name, ' ') {
			return fmt.Errorf("task [%d]: name '%s': %w", i, t.Name, ErrHasSpaces)
		}
	}

	for i, n := range config.Notifiers {
		if strings.ContainsRune(n.Name, ' ') {
			return fmt.Errorf("notifier [%d]: name '%s': %w", i, n.Name, ErrHasSpaces)
		}
	}

	for i, g := range config.Gates {
		if strings.ContainsRune(g.Name, ' ') {
			return fmt.Errorf("gate [%d]: name '%s': %w", i, g.Name, ErrHasSpaces)
		}
	}

	return nil
}

func (c Config) validateFrequencySeconds() error {
	if c.Defaults.FrequencySeconds < 1 {
		return fmt.Errorf("default: frequency '%d': %w", c.Defaults.FrequencySeconds, ErrLessThan1)
	}

	if c.Defaults.RetryFrequencySeconds < 1 {
		return fmt.Errorf("default: retry_frequency '%d': %w", c.Defaults.RetryFrequencySeconds, ErrLessThan1)
	}

	for i, t := range config.Tasks {
		if t.FrequencySeconds < 1 {
			return fmt.Errorf("task [%d]: freqency '%d': %w", i, t.FrequencySeconds, ErrLessThan1)
		}

		if t.RetryFrequencySeconds < 1 {
			return fmt.Errorf("task [%d]: retry_freqency '%d': %w", i, t.RetryFrequencySeconds, ErrLessThan1)
		}
	}

	return nil
}

func (c Config) validateTimeoutSeconds() error {
	if c.Defaults.TimeoutSeconds < 1 {
		return fmt.Errorf("default: timeout_seconds '%d': %w", c.Defaults.TimeoutSeconds, ErrLessThan1)
	}

	if c.Defaults.TimeoutSeconds > c.Defaults.RetryFrequencySeconds {
		return fmt.Errorf("default: timeout_seconds '%d': %w", c.Defaults.TimeoutSeconds, ErrTimeoutSlowerThanRetry)
	}

	for i, t := range config.Tasks {
		if t.TimeoutSeconds < 1 {
			return fmt.Errorf("task [%d]: timeout_seconds '%d': %w", i, t.TimeoutSeconds, ErrLessThan1)
		}

		if t.TimeoutSeconds > t.RetryFrequencySeconds {
			return fmt.Errorf("task [%d]: timeout_seconds '%d': %w", i, t.RetryFrequencySeconds, ErrTimeoutSlowerThanRetry)
		}
	}

	return nil
}

func (c Config) validatePriority() error {
	if c.Defaults.Priority < 1 {
		return fmt.Errorf("default: priority '%d': %w", c.Defaults.Priority, ErrLessThan1)
	}
	if c.Defaults.Priority > 99 {
		return fmt.Errorf("default: priority '%d': %w", c.Defaults.Priority, ErrGreaterThan99)
	}

	for i, t := range config.Tasks {
		if t.Priority < 1 {
			return fmt.Errorf("task [%d]: priority '%d': %w", i, t.Priority, ErrLessThan1)
		}
		if t.Priority > 99 {
			return fmt.Errorf("task [%d]: priority '%d': %w", i, t.Priority, ErrGreaterThan99)
		}
	}

	return nil
}
