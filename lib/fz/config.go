package fz

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pelletier/go-toml"
)

var DAEMON_START_TIME = time.Now()

var config Config

func ReadConfig(f string) Config {
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

	if config.Defaults.Retries == 0 {
		config.Defaults.Retries = DEFAULT_RETRIES
	}

	if config.Defaults.FrequencySeconds == 0 {
		config.Defaults.FrequencySeconds = DEFAULT_FREQUENCY_SECONDS
	}

	if config.Defaults.RetryFrequencySeconds == 0 {
		config.Defaults.RetryFrequencySeconds = config.Defaults.FrequencySeconds
	}

	if config.Defaults.TimeoutSeconds == 0 {
		config.Defaults.TimeoutSeconds = DEFAULT_TIMEOUT_SECONDS
	}

	if config.Defaults.Priority == 0 {
		config.Defaults.Priority = DEFAULT_PRIORITY
	}

	for i, t := range config.Tasks {
		if t.Retries == 0 {
			config.Tasks[i].Retries = config.Defaults.Retries
		}

		if t.TimeoutSeconds == 0 {
			config.Tasks[i].TimeoutSeconds = config.Defaults.TimeoutSeconds
		}

		if t.FrequencySeconds == 0 {
			config.Tasks[i].FrequencySeconds = config.Defaults.FrequencySeconds
		}

		if t.RetryFrequencySeconds == 0 {
			config.Tasks[i].RetryFrequencySeconds = config.Defaults.RetryFrequencySeconds
		}

		if t.Priority == 0 {
			config.Tasks[i].Priority = config.Defaults.Priority
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

		if len(t.ErrorNotifierNames) == 0 {
			config.Tasks[i].ErrorNotifierNames = config.Defaults.ErrorNotifierNames
		}

		// start the history in an unknown state
		config.Tasks[i].History = 0b10
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

// Validate the configuration
func (c Config) Validate() error {
	if err := c.validateNotifiersExist(); err != nil {
		return err
	}

	if err := c.validateGatesExist(); err != nil {
		return err
	}

	if err := c.validateCommandsExist(); err != nil {
		return err
	}

	if err := c.validateName(); err != nil {
		return err
	}

	if err := c.validateFrequencySeconds(); err != nil {
		return err
	}

	if err := c.validateTimeoutSeconds(); err != nil {
		return err
	}

	if err := c.validatePriority(); err != nil {
		return err
	}

	for i, t := range config.Tasks {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("task [%d]: %w", i, err)
		}
	}

	return nil
}

// Find and return a Notifier by its name.
func (c Config) GetNotifierByName(name string) *Notifier {
	for i, n := range c.Notifiers {
		if n.Name == name {
			return &c.Notifiers[i]
		}
	}

	return nil
}

// Find and return a Gate by its name.
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

func (c Config) ErrorNotification() {
	//for _, n := range c.ErrorNotifiers {
	//	NotifyCh <- Notification{n, t}
	//}
}

///////////////////////////////////////////////////////////////////////////////
// Private Methods

func (c Config) validateNotifiersExist() error {
	for _, n := range c.Defaults.NotifierNames {
		if c.GetNotifierByName(n) == nil {
			return fmt.Errorf("default: notifier %s: %w", n, ErrNotExist)
		}
	}

	for _, n := range c.Defaults.ErrorNotifierNames {
		if c.GetNotifierByName(n) == nil {
			return fmt.Errorf("default: error_notifier %s: %w", n, ErrNotExist)
		}
	}

	for _, t := range c.Tasks {
		for i, n := range t.NotifierNames {
			if c.GetNotifierByName(n) == nil {
				return fmt.Errorf("task [%d]: notifier '%s': %w", i, n, ErrNotExist)
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
		if _, err := os.Stat(filepath.Join(c.Directory, t.Command)); os.IsNotExist(err) {
			return fmt.Errorf("task [%d]: command '%s': %w", i, t.Command, ErrCommandNotExist)
		}
	}

	for i, n := range config.Notifiers {
		if _, err := os.Stat(filepath.Join(c.Directory, n.Command)); os.IsNotExist(err) {
			return fmt.Errorf("notifier [%d]: command '%s': %w", i, n.Command, ErrCommandNotExist)
		}
	}

	for i, g := range config.Gates {
		if _, err := os.Stat(filepath.Join(c.Directory, g.Command)); os.IsNotExist(err) {
			return fmt.Errorf("gate [%d]: command '%s': %w", i, g.Command, ErrCommandNotExist)
		}
	}

	return nil
}

func (c Config) validateName() error {
	re := regexp.MustCompile(`^[a-z0-9_]+$`)

	for i, n := range config.Notifiers {
		if !re.Match([]byte(n.Name)) {
			return fmt.Errorf("notifier [%d]: name '%s': %w", i, n.Name, ErrInvalidName)
		}
	}

	for i, g := range config.Gates {
		if !re.Match([]byte(g.Name)) {
			return fmt.Errorf("notifier [%d]: name '%s': %w", i, g.Name, ErrInvalidName)
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

	return nil
}

func (c Config) validateTimeoutSeconds() error {
	if c.Defaults.TimeoutSeconds < 1 {
		return fmt.Errorf("default: timeout_seconds '%d': %w", c.Defaults.TimeoutSeconds, ErrLessThan1)
	}

	if c.Defaults.TimeoutSeconds > c.Defaults.RetryFrequencySeconds {
		return fmt.Errorf("default: timeout_seconds '%d': %w", c.Defaults.TimeoutSeconds, ErrTimeoutSlowerThanRetry)
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

	return nil
}
