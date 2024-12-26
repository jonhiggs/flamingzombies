package fz

import "errors"

var ErrCommandNotExist = errors.New("command does not exist")
var ErrNoSpaces = errors.New("cannot contain spaces")
var ErrNotExist = errors.New("does not exist")
