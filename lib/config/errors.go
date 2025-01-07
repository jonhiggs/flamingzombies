package config

import "errors"

var ErrCommandNotExist = errors.New("command does not exist")
var ErrGreaterThan99 = errors.New("cannot be greater than 99")
var ErrInvalidName = errors.New("only alphanumeric, underscore and colon characters are allowed")
var ErrInvalidPermissions = errors.New("invalid permissions")
var ErrLessThan1 = errors.New("cannot be less than 1")
var ErrNotExist = errors.New("does not exist")
var ErrRetriesSlowerThanFrequency = errors.New("retry_frequency must be less than frequency")
var ErrTimeout = errors.New("timeout exceeded")
var ErrTimeoutSlowerThanRetry = errors.New("timeout must not be longer than the retry interval")
