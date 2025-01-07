package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	fhA, _ := os.Open("./examples/task_simple.toml")

	var tests = []struct {
		fh           *os.File
		wantDir      string
		wantLogFile  string
		wantLogLevel string
		wantDef      defaultConfig
		wantErr      error
	}{
		{ // A
			fh:           fhA,
			wantDir:      "/usr/lib/flamingzombies",
			wantLogFile:  "-",
			wantLogLevel: "info",
			wantDef: defaultConfig{
				Retries:        5,
				TimeoutSeconds: 1,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.fh.Name()), func(t *testing.T) {
			err := Load(tt.fh)
			if err != nil {
				panic(fmt.Errorf("load: %w", err))
			}

			if def.Retries != tt.wantDef.Retries {
				t.Errorf("default retries got: %d, want: %d", def.Retries, tt.wantDef.Retries)
			}

			if def.TimeoutSeconds != tt.wantDef.TimeoutSeconds {
				t.Errorf("default timeout got: %d, want: %d", def.TimeoutSeconds, tt.wantDef.TimeoutSeconds)
			}

			if err != tt.wantErr {
				t.Errorf("\ngot: %s\nwant: %s\n", err, tt.wantErr)
			}

			if Directory != tt.wantDir {
				t.Errorf("Directory got: %s, want: %s", Directory, tt.wantDir)
			}

			if LogFile != tt.wantLogFile {
				t.Errorf("LogFile got: %s, want: %s", LogFile, tt.wantLogFile)
			}

			if LogLevel != tt.wantLogLevel {
				t.Errorf("LogLevel got: %s, want: %s", LogLevel, tt.wantLogLevel)
			}

			if len(Tasks) != 1 {
				t.Errorf("task count got: %d, want: %d", len(Tasks), 1)
			}

			if len(Gates) != 2 {
				t.Errorf("gate count got: %d, want: %d", len(Gates), 2)
			}

			if len(Notifiers) != 1 {
				t.Errorf("gate count got: %d, want: %d", len(Notifiers), 1)
			}

			// unset the global state
			def = defaultConfig{}
			Tasks = []Task{}
			Gates = []Gate{}
			Notifiers = []Notifier{}
			Directory = ""
		})
	}
}

func TestExtractTomlObjectsBasic(t *testing.T) {
	f, _ := os.Open("./examples/basic.toml")
	b, err := PreProcess(f, []*os.File{})
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	t.Run("default", func(t *testing.T) {
		got := extractTomlOjbects("default", b)

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
		}

		gotLines := strings.Split(string(got[0]), "\n")

		if len(gotLines) != 17 {
			t.Errorf("got: %d, want: %d", len(gotLines), 17)
		}

		if gotLines[0] != "retries = 5" {
			t.Errorf("got: %s, want: %s", gotLines[0], "retries = 5")
		}

	})

	t.Run("task", func(t *testing.T) {
		got := extractTomlOjbects("task", b)

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
		}

		if len(strings.Split(string(got[0]), "\n")) != 9 {
			t.Errorf("got: %d, want: %d", len(strings.Split(string(got[0]), "\n")), 9)
		}
	})

	t.Run("gate", func(t *testing.T) {
		got := extractTomlOjbects("gate", b)

		if len(got) != 6 {
			t.Errorf("got: %d, want: %d", len(got), 6)
		}
	})

	t.Run("notifier", func(t *testing.T) {
		got := extractTomlOjbects("notifier", b)

		if len(got) != 3 {
			t.Errorf("got: %d, want: %d", len(got), 3)
		}
	})
}

func TestTasksFromToml(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		b := []byte{}
		got, err := tasksFromToml(b, defaultConfig{})

		if err != nil {
			t.Errorf("got: %v, want: %v", err, nil)
		}

		if len(got) != 0 {
			t.Errorf("got: %d, want: %d", len(got), 0)
		}
	})

	t.Run("simple task", func(t *testing.T) {
		fh, _ := os.Open("./examples/task_simple.toml")
		b, err := PreProcess(fh, []*os.File{})
		if err != nil {
			panic(fmt.Errorf("preProcess: %w", err))
		}

		d, err := defaultConfigFromToml(extractTomlOjbects("default", b)[0])
		if err != nil {
			panic(fmt.Errorf("tomlDefaultConfig: %w", err))
		}
		got, err := tasksFromToml(b, d)

		if err != nil {
			t.Errorf("got: %s, want: %v", err, nil)
		}

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
		}

		if got[0].Name != "simple task" {
			t.Errorf("got: %s, want: %s", got[0].Name, "simple task")
		}

		if got[0].Description != "this is a simple task" {
			t.Errorf("got: %s, want: %s", got[0].Description, "this is a simple task")
		}

		if got[0].Command != "task/command" {
			t.Errorf("got: %s, want: %s", got[0].Command, "task/command")
		}

		if got[0].FrequencySeconds != 20 {
			t.Errorf("got: %d, want: %d", got[0].FrequencySeconds, 20)
		}

		if got[0].RetryFrequencySeconds != 20 {
			t.Errorf("got: %d, want: %d", got[0].RetryFrequencySeconds, 20)
		}

		if got[0].Priority != 3 {
			t.Errorf("got: %d, want: %d", got[0].Priority, 3)
		}

		envs := []string{
			"EXTRA=123",
			"SNMP_COMMUNITY=public",
			"SNMP_VERSION=2c",
			"EMAIL_FROM=fz@example",
		}

		if fmt.Sprintf("%s", got[0].Envs) != fmt.Sprintf("%s", envs) {
			t.Errorf("got: %s, want: %s", got[0].Envs, envs)
		}

		// unset global state
		def = defaultConfig{}
	})

	t.Run("no retries", func(t *testing.T) {
		fh, _ := os.Open("./examples/no_retries.toml")
		b, err := PreProcess(fh, []*os.File{})
		if err != nil {
			panic(fmt.Errorf("preProcess: %w", err))
		}

		d, err := defaultConfigFromToml(extractTomlOjbects("default", b)[0])
		if err != nil {
			panic(fmt.Errorf("tomlDefaultConfig: %w", err))
		}
		got, err := tasksFromToml(b, d)

		if err != nil {
			t.Errorf("got: %s, want: %v", err, nil)
		}

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
		}

		if got[0].Name != "no_retries" {
			t.Errorf("got: %s, want: %s", got[0].Name, "no_retries")
		}

		if got[0].Retries != 0 {
			t.Errorf("got: %d, want: %d", got[0].Retries, 0)
		}

		if got[0].RetryFrequencySeconds != 0 {
			t.Errorf("got: %d, want: %d", got[0].RetryFrequencySeconds, 0)
		}

		// unset global state
		def = defaultConfig{}
	})
}

func TestGatesFromToml(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		b := []byte{}
		got, err := gatesFromToml(b, defaultConfig{})

		if err != nil {
			t.Errorf("got: %v, want: %v", err, nil)
		}

		if len(got) != 0 {
			t.Errorf("got: %d, want: %d", len(got), 0)
		}
	})

	t.Run("simple task", func(t *testing.T) {
		fh, _ := os.Open("./examples/task_simple.toml")
		b, err := PreProcess(fh, []*os.File{})
		if err != nil {
			panic(fmt.Errorf("preProcess: %w", err))
		}

		d, err := defaultConfigFromToml(extractTomlOjbects("default", b)[0])
		if err != nil {
			panic(fmt.Errorf("tomlDefaultConfig: %w", err))
		}
		got, err := gatesFromToml(b, d)

		if err != nil {
			t.Errorf("got: %s, want: %v", err, nil)
		}

		if len(got) != 2 {
			t.Errorf("got: %d, want: %d", len(got), 2)
		}

		if got[0].Name() != "to_failed" {
			t.Errorf("got: %s, want: %s", got[0].Name(), "to_failed")
		}

		if got[1].Name() != "to_ok" {
			t.Errorf("got: %s, want: %s", got[0].Name(), "to_ok")
		}

		if got[0].Command() != "gate/to_state" {
			t.Errorf("got: %s, want: %s", got[0].Command(), "gate/to_state")
		}

		if got[1].Command() != "gate/to_state" {
			t.Errorf("got: %s, want: %s", got[1].Command(), "gate/to_state")
		}

		if fmt.Sprintf("%v", got[0].Args()) != fmt.Sprintf("%v", []string{"fail"}) {
			t.Errorf("got: %v, want: %v", got[0].Args(), []string{"fail"})
		}

		if fmt.Sprintf("%v", got[1].Args()) != fmt.Sprintf("%v", []string{"ok"}) {
			t.Errorf("got: %v, want: %v", got[1].Args(), []string{"ok"})
		}

		envs := []string{
			"SNMP_COMMUNITY=public",
			"SNMP_VERSION=2c",
			"EMAIL_FROM=fz@example",
		}

		if fmt.Sprintf("%s", got[0].Environment()) != fmt.Sprintf("%s", envs) {
			t.Errorf("got: %s, want: %s", got[0].Environment(), envs)
		}

		if fmt.Sprintf("%s", got[1].Environment()) != fmt.Sprintf("%s", envs) {
			t.Errorf("got: %s, want: %s", got[1].Environment(), envs)
		}
	})
}

func TestNotifiersFromToml(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		b := []byte{}
		got, err := notifiersFromToml(b, defaultConfig{})

		if err != nil {
			t.Errorf("got: %v, want: %v", err, nil)
		}

		if len(got) != 0 {
			t.Errorf("got: %d, want: %d", len(got), 0)
		}
	})

	t.Run("simple task", func(t *testing.T) {
		fh, _ := os.Open("./examples/task_simple.toml")
		b, err := PreProcess(fh, []*os.File{})
		if err != nil {
			panic(fmt.Errorf("preProcess: %w", err))
		}

		d, err := defaultConfigFromToml(extractTomlOjbects("default", b)[0])
		if err != nil {
			panic(fmt.Errorf("tomlDefaultConfig: %w", err))
		}
		got, err := notifiersFromToml(b, d)

		if err != nil {
			t.Errorf("got: %s, want: %v", err, nil)
		}

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
		}

		if got[0].Name() != "mailer" {
			t.Errorf("got: %s, want: %s", got[0].Name(), "mailer")
		}

		if got[0].Command() != "notifier/email" {
			t.Errorf("got: %s, want: %s", got[0].Command(), "notifier/email")
		}

		envs := []string{
			"EMAIL_ADDRESS=root@example",
			"SNMP_COMMUNITY=public",
			"SNMP_VERSION=2c",
			"EMAIL_FROM=fz@example",
		}

		if fmt.Sprintf("%s", got[0].Environment()) != fmt.Sprintf("%s", envs) {
			t.Errorf("got: %s, want: %s", got[0].Environment(), envs)
		}
	})
}
