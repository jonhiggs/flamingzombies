package run

import (
	"io"
	"os/exec"
	"time"
)

// The result of a command evaluation.
type Result struct {
	StdoutBytes []byte
	StderrBytes []byte
	Duration    time.Duration
	ExitCode    int
	Err         error
	TraceID     string
}

func Cmd(cmd *exec.Cmd, traceID string) Result {
	var r Result
	r.TraceID = traceID

	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	startTime := time.Now()
	err := cmd.Start()
	if err != nil {
		r.Err = err
		r.ExitCode = -1
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
