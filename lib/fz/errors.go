package fz

import "errors"

var ErrCommandNotExist = errors.New("command does not exist")
var ErrHasSpaces = errors.New("cannot contain spaces")
var ErrNotExist = errors.New("does not exist")
var ErrTooFrequent = errors.New("cannot not be more frequent than every second")
