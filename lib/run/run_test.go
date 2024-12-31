package run

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestRunStart(t *testing.T) {
	var tests = []struct {
		name string
		cmd  Cmd
		want Result
	}{
		{
			name: "true",
			cmd: Cmd{
				Command: "true",
				Args:    []string{},
				Envs:    []string{},
				Dir:     "/",
				TraceID: "0",
				Timeout: 1 * time.Second,
			},
			want: Result{
				StdoutBytes: []byte(""),
				StderrBytes: []byte(""),
				ExitCode:    0,
				Err:         nil,
				TraceID:     "0",
			},
		},
		{
			name: "false",
			cmd: Cmd{
				Command: "false",
				Args:    []string{},
				Envs:    []string{},
				Dir:     "/",
				TraceID: "1",
				Timeout: 1 * time.Second,
			},
			want: Result{
				StdoutBytes: []byte(""),
				StderrBytes: []byte(""),
				ExitCode:    1,
				Err:         nil,
				TraceID:     "1",
			},
		},
		{
			name: "hello",
			cmd: Cmd{
				Command: "echo",
				Args:    []string{"-n", "hello"},
				Envs:    []string{},
				Dir:     "/",
				TraceID: "2",
				Timeout: 1 * time.Second,
			},
			want: Result{
				StdoutBytes: []byte("hello"),
				StderrBytes: []byte(""),
				ExitCode:    0,
				Err:         nil,
				TraceID:     "2",
			},
		},
		{
			name: "error_hello",
			cmd: Cmd{
				Command: "test_commands/error_hello",
				Args:    []string{"-n", "hello"},
				Envs:    []string{},
				Dir:     "./",
				TraceID: "3",
				Timeout: 1 * time.Second,
			},
			want: Result{
				StdoutBytes: []byte(""),
				StderrBytes: []byte("hello"),
				ExitCode:    0,
				Err:         nil,
				TraceID:     "3",
			},
		},
		{
			name: "test_env_bound_X",
			cmd: Cmd{
				Command: "test_commands/test_env",
				Args:    []string{"X"},
				Envs:    []string{"X=yes"},
				Dir:     "./",
				TraceID: "4",
				Timeout: 1 * time.Second,
			},
			want: Result{
				StdoutBytes: []byte("yes"),
				StderrBytes: []byte(""),
				ExitCode:    0,
				Err:         nil,
				TraceID:     "4",
			},
		},
		{
			name: "test_env_unbound_X",
			cmd: Cmd{
				Command: "test_commands/test_env",
				Args:    []string{"X"},
				Envs:    []string{""},
				Dir:     "./",
				TraceID: "5",
				Timeout: 1 * time.Second,
			},
			want: Result{
				StdoutBytes: []byte(""),
				StderrBytes: []byte("test_commands/test_env: line 3: !1: unbound variable"),
				ExitCode:    1,
				Err:         nil,
				TraceID:     "5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.name), func(t *testing.T) {
			got := tt.cmd.Start()

			if got.ExitCode != tt.want.ExitCode {
				t.Errorf("exit code: got: %d, want: %d", got.ExitCode, tt.want.ExitCode)
			}

			if got.TraceID != tt.want.TraceID {
				t.Errorf("trace id: got: %s, want: %s", got.TraceID, tt.want.TraceID)
			}

			if got.Stdout() != tt.want.Stdout() {
				t.Errorf("stdout: got: %s, want: %s", got.Stdout(), tt.want.Stdout())
			}

			if got.Stderr() != tt.want.Stderr() {
				t.Errorf("stderr: got: %s, want: %s", got.Stderr(), tt.want.Stderr())
			}

			if got.Duration == 0 {
				t.Errorf("duration: got %d, want something greater than 0", got.Duration)
			}

			if got.Err != nil {
				t.Errorf("error: got %s, want nil", got.Err)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		cmd := Cmd{
			Command: "soeijspeifjspiefjsoijef",
			Timeout: 1 * time.Second,
		}

		want := Result{
			ExitCode: -1,
			Err:      exec.ErrNotFound,
		}

		got := cmd.Start()

		if errors.Unwrap(got.Err) != want.Err {
			t.Errorf("error: got %s, want %s", errors.Unwrap(got.Err), want.Err)
		}

		if got.ExitCode != want.ExitCode {
			t.Errorf("exit code: got: %d, want: %d", got.ExitCode, want.ExitCode)
		}
	})

	t.Run("no exec bit", func(t *testing.T) {
		cmd := Cmd{
			Command: "test_commands/no_exec_bit",
			Timeout: 1 * time.Second,
		}

		want := Result{
			ExitCode: -1,
			Err:      syscall.EACCES,
		}

		got := cmd.Start()

		if errors.Unwrap(got.Err) != want.Err {
			t.Errorf("error: got %s, want %s", errors.Unwrap(got.Err), want.Err)
		}

		if got.ExitCode != want.ExitCode {
			t.Errorf("exit code: got: %d, want: %d", got.ExitCode, want.ExitCode)
		}
	})
}
