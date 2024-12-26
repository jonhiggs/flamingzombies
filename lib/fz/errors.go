package fz

import "errors"

var ErrCommandNotExist = errors.New("command does not exist")
var ErrHasSpaces = errors.New("cannot contain spaces")
var ErrNotExist = errors.New("does not exist")
var ErrLessThan1 = errors.New("cannot be less than 1")
var ErrTimeoutSlowerThanRetry = errors.New("timeout must not be longer than the retry interval")
var ErrGreaterThan99 = errors.New("cannot be greater than 99")
