package fz

import (
	"time"
)

var DAEMON_START_TIME = time.Now()

//// Validate the configuration
//func (c Config) Validate() error {
//	if err := c.validateNotifiersExist(); err != nil {
//		return err
//	}
//
//	if err := c.validateGatesExist(); err != nil {
//		return err
//	}
//
//	if err := c.validateCommandsExist(); err != nil {
//		return err
//	}
//
//	if err := c.validateName(); err != nil {
//		return err
//	}
//
//	if err := c.validateFrequencySeconds(); err != nil {
//		return err
//	}
//
//	if err := c.validateTimeoutSeconds(); err != nil {
//		return err
//	}
//
//	if err := c.validatePriority(); err != nil {
//		return err
//	}
//
//	for i, t := range cfg.Tasks {
//		if err := t.Validate(); err != nil {
//			return fmt.Errorf("task %s [%d]: %w", t.Name, i, err)
//		}
//	}
//
//	return nil
//}
//
//// Find and return a Task by its name.
//func (c Config) GetTaskByName(name string) *Task {
//	for i, t := range c.Tasks {
//		if t.Name == name {
//			return &c.Tasks[i]
//		}
//	}
//
//	return nil
//}
//
//// Find and return a Notifier by its name.
//func (c Config) GetNotifierByName(name string) *Notifier {
//	for i, n := range c.Notifiers {
//		if n.Name == name {
//			return &c.Notifiers[i]
//		}
//	}
//
//	return nil
//}
//
//// Find and return a Gate by its name.
//func (c Config) GetGateByName(name string) *Gate {
//	for i, g := range c.Gates {
//		if g.Name == name {
//			return &c.Gates[i]
//		}
//	}
//
//	return nil
//}
//
//func (c Config) ErrorNotification() {
//	//for _, n := range c.ErrorNotifiers {
//	//	NotifyCh <- Notification{n, t}
//	//}
//}
//
/////////////////////////////////////////////////////////////////////////////////
//// Private Methods
//
//func (c Config) validateNotifiersExist() error {
//	for _, n := range c.Defaults.NotifierNames {
//		if c.GetNotifierByName(n) == nil {
//			return fmt.Errorf("default: notifier %s: %w", n, ErrNotExist)
//		}
//	}
//
//	for _, n := range c.Defaults.ErrorNotifierNames {
//		if c.GetNotifierByName(n) == nil {
//			return fmt.Errorf("default: error_notifier %s: %w", n, ErrNotExist)
//		}
//	}
//
//	for _, t := range c.Tasks {
//		for i, n := range t.NotifierNames {
//			if c.GetNotifierByName(n) == nil {
//				return fmt.Errorf("task %s [%d]: notifier '%s': %w", t.Name, i, n, ErrNotExist)
//			}
//		}
//	}
//
//	return nil
//}
//
//func (c Config) validateGatesExist() error {
//	for _, n := range c.Notifiers {
//		for i, gs := range n.GateSetStrings {
//			for ii, g := range gs {
//				if c.GetGateByName(g) == nil {
//					return fmt.Errorf("gate [%d][%d]: %w", i, ii, ErrNotExist)
//				}
//			}
//		}
//	}
//
//	return nil
//}
//
//func (c Config) validateCommandsExist() error {
//	for i, t := range cfg.Tasks {
//		cmd := filepath.Join(c.Directory, t.Command)
//		Logger.Debug("checking command", "cmd", cmd)
//
//		if _, err := os.Stat(cmd); os.IsNotExist(err) {
//			return fmt.Errorf("task %s [%d]: command '%s': %w", t.Name, i, t.Command, ErrCommandNotExist)
//		}
//	}
//
//	for i, n := range cfg.Notifiers {
//		cmd := filepath.Join(c.Directory, n.Command)
//		Logger.Debug("checking command", "cmd", cmd)
//
//		if _, err := os.Stat(cmd); os.IsNotExist(err) {
//			return fmt.Errorf("notifier [%d]: command '%s': %w", i, n.Command, ErrCommandNotExist)
//		}
//	}
//
//	for i, g := range cfg.Gates {
//		cmd := filepath.Join(c.Directory, g.Command)
//		Logger.Debug("checking command", "cmd", cmd)
//
//		if _, err := os.Stat(cmd); os.IsNotExist(err) {
//			return fmt.Errorf("gate [%d]: command '%s': %w", i, g.Command, ErrCommandNotExist)
//		}
//	}
//
//	return nil
//}
//
//func (c Config) validateName() error {
//	re := regexp.MustCompile(`^[a-z0-9_:]+$`)
//
//	for i, n := range cfg.Notifiers {
//		if !re.Match([]byte(n.Name)) {
//			return fmt.Errorf("notifier [%d]: name '%s': %w", i, n.Name, ErrInvalidName)
//		}
//	}
//
//	for i, g := range cfg.Gates {
//		if !re.Match([]byte(g.Name)) {
//			return fmt.Errorf("notifier [%d]: name '%s': %w", i, g.Name, ErrInvalidName)
//		}
//	}
//
//	return nil
//}
//
//func (c Config) validateFrequencySeconds() error {
//	if c.Defaults.FrequencySeconds < 1 {
//		return fmt.Errorf("default: frequency '%d': %w", c.Defaults.FrequencySeconds, ErrLessThan1)
//	}
//
//	if c.Defaults.RetryFrequencySeconds < 1 {
//		return fmt.Errorf("default: retry_frequency '%d': %w", c.Defaults.RetryFrequencySeconds, ErrLessThan1)
//	}
//
//	return nil
//}
//
//func (c Config) validateTimeoutSeconds() error {
//	if c.Defaults.TimeoutSeconds < 1 {
//		return fmt.Errorf("default: timeout_seconds '%d': %w", c.Defaults.TimeoutSeconds, ErrLessThan1)
//	}
//
//	if c.Defaults.TimeoutSeconds > c.Defaults.RetryFrequencySeconds {
//		return fmt.Errorf("default: timeout_seconds '%d': %w", c.Defaults.TimeoutSeconds, ErrTimeoutSlowerThanRetry)
//	}
//
//	return nil
//}
//
//func (c Config) validatePriority() error {
//	if c.Defaults.Priority < 1 {
//		return fmt.Errorf("default: priority '%d': %w", c.Defaults.Priority, ErrLessThan1)
//	}
//	if c.Defaults.Priority > 99 {
//		return fmt.Errorf("default: priority '%d': %w", c.Defaults.Priority, ErrGreaterThan99)
//	}
//
//	return nil
//}
