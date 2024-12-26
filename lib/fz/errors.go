package fz

import "errors"

var ErrCommandNotExist = errors.New("command does not exist")
var ErrInvalidName = errors.New("charactors must be alphanumeric or underscore")
var ErrNotExist = errors.New("does not exist")
var ErrLessThan1 = errors.New("cannot be less than 1")
var ErrTimeoutSlowerThanRetry = errors.New("timeout must not be longer than the retry interval")
var ErrGreaterThan99 = errors.New("cannot be greater than 99")
