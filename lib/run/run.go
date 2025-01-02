package run

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

const GRACE_TIME = time.Duration(500) * time.Millisecond

type Cmd struct {
	Command string
	Args    []string
	Envs    []string
	Dir     string
	TraceID string
	Timeout time.Duration
}

// The result of a command evaluation.
type Result struct {
	StdoutBytes []byte
	StderrBytes []byte
	Duration    time.Duration
	ExitCode    int
	Err         error
	TraceID     string
}

func (c Cmd) Start() Result {
	var r Result

	r.TraceID = c.TraceID
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout())
	defer cancel()

	cmd := exec.CommandContext(ctx, c.Command, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = c.Envs

	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		r.Err = err
		r.ExitCode = 255
		return r
	}
	r.StdoutBytes, _ = io.ReadAll(stdout)
	r.StderrBytes, _ = io.ReadAll(stderr)
	err = cmd.Wait()
	r.Duration = time.Now().Sub(startTime)

	if err != nil {
		v, ok := err.(*exec.ExitError)
		if ok {
			r.ExitCode = v.ExitCode()
		} else {
			r.Err = err
		}
	}

	return r
}

// There will always be a little descrepency between the Go Timeout and the
// actual time remaining once the process has forked and has began executing
// the user-code. Adding GRACE_TIME to Cmd.Timeout ensures that the running
// process will always get at least the requested Timeout of running time.
func (c Cmd) timeout() time.Duration {
	return GRACE_TIME + c.Timeout
}

func (r Result) Stdout() string {
	return strings.TrimSuffix(string(r.StdoutBytes), "\n")
}

func (r Result) Stderr() string {
	return strings.TrimSuffix(string(r.StderrBytes), "\n")
}
