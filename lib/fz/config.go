package fz

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/preprocessor"
	"github.com/pelletier/go-toml"
)

var DAEMON_START_TIME = time.Now()
var cfg Config

// Make configuration available in fz.Configuration
func ReadConfig(f, dir, logFile, logLevel string) *Config {

	fh, err := os.Open(f)
	if err != nil {
		log.Fatal(fmt.Errorf("reading config: %w", err))
	}
	defer fh.Close()

	//b, err := io.ReadAll(file)
	//if err != nil {
	//	Fatal(fmt.Sprintf("Error reading the configuration file '%s'\n", f), fmt.Sprint(err))
	//}

	b, err := preprocessor.Run(fh, []*os.File{})
	if err != nil {
		log.Fatal(fmt.Errorf("preprocessing config: %w", err))
	}

	err = toml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("parsing config: %w", err))
	}

	// Replace the values set by TOML with those supplied to func.
	if dir != "" {
		cfg.Directory = dir
	}

	if logFile != "" {
		cfg.LogFile = logFile
	}

	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if cfg.Defaults.Retries == 0 {
		cfg.Defaults.Retries = DEFAULT_RETRIES
	}

	if cfg.Defaults.FrequencySeconds == 0 {
		cfg.Defaults.FrequencySeconds = DEFAULT_FREQUENCY_SECONDS
	}

	if cfg.Defaults.RetryFrequencySeconds == 0 {
		cfg.Defaults.RetryFrequencySeconds = cfg.Defaults.FrequencySeconds
	}

	if cfg.Defaults.TimeoutSeconds == 0 {
		cfg.Defaults.TimeoutSeconds = DEFAULT_TIMEOUT_SECONDS
	}

	if cfg.Defaults.Priority == 0 {
		cfg.Defaults.Priority = DEFAULT_PRIORITY
	}

	for i, t := range cfg.Tasks {
		if t.Retries == 0 {
			cfg.Tasks[i].Retries = cfg.Defaults.Retries
		}

		if t.TimeoutSeconds == 0 {
			cfg.Tasks[i].TimeoutSeconds = cfg.Defaults.TimeoutSeconds
		}

		if t.FrequencySeconds == 0 {
			cfg.Tasks[i].FrequencySeconds = cfg.Defaults.FrequencySeconds
		}

		if t.RetryFrequencySeconds == 0 {
			cfg.Tasks[i].RetryFrequencySeconds = cfg.Defaults.RetryFrequencySeconds
		}

		if t.Priority == 0 {
			cfg.Tasks[i].Priority = cfg.Defaults.Priority
		}

		if len(t.ErrorBody) == 0 {
			cfg.Tasks[i].ErrorBody = "The task has entered an error state"
		}
		if len(t.RecoverBody) == 0 {
			cfg.Tasks[i].RecoverBody = "The task has recovered from an error state"
		}

		if len(t.NotifierNames) == 0 {
			cfg.Tasks[i].NotifierNames = cfg.Defaults.NotifierNames
		}

		if len(t.ErrorNotifierNames) == 0 {
			cfg.Tasks[i].ErrorNotifierNames = cfg.Defaults.ErrorNotifierNames
		}

		// start the history in an unknown state
		cfg.Tasks[i].History = 0b10
	}

	return &cfg
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

	for i, t := range cfg.Tasks {
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
	r := [][]*Gate{}

	n := c.GetNotifierByName(notifierName)
	for _, gateSet := range n.GateSets {
		gs := []*Gate{}
		for _, gateName := range gateSet {
			gs = append(gs, cfg.GetGateByName(gateName))
		}
		r = append(r, gs)
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
	for i, t := range cfg.Tasks {
		cmd := filepath.Join(c.Directory, t.Command)
		Logger.Debug("checking command", "cmd", cmd)

		if _, err := os.Stat(cmd); os.IsNotExist(err) {
			return fmt.Errorf("task [%d]: command '%s': %w", i, t.Command, ErrCommandNotExist)
		}
	}

	for i, n := range cfg.Notifiers {
		cmd := filepath.Join(c.Directory, n.Command)
		Logger.Debug("checking command", "cmd", cmd)

		if _, err := os.Stat(cmd); os.IsNotExist(err) {
			return fmt.Errorf("notifier [%d]: command '%s': %w", i, n.Command, ErrCommandNotExist)
		}
	}

	for i, g := range cfg.Gates {
		cmd := filepath.Join(c.Directory, g.Command)
		Logger.Debug("checking command", "cmd", cmd)

		if _, err := os.Stat(cmd); os.IsNotExist(err) {
			return fmt.Errorf("gate [%d]: command '%s': %w", i, g.Command, ErrCommandNotExist)
		}
	}

	return nil
}

func (c Config) validateName() error {
	re := regexp.MustCompile(`^[a-z0-9_:]+$`)

	for i, n := range cfg.Notifiers {
		if !re.Match([]byte(n.Name)) {
			return fmt.Errorf("notifier [%d]: name '%s': %w", i, n.Name, ErrInvalidName)
		}
	}

	for i, g := range cfg.Gates {
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
