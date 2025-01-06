package config

import (
	"fmt"
	"os"
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
				panic(fmt.Errorf("%w", err))
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

			// unset the global state
			def = defaultConfig{}
			Tasks = []Task{}
			Directory = ""
		})
	}
}

func TestExtractTomlObjects(t *testing.T) {
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
	})

	t.Run("task", func(t *testing.T) {
		got := extractTomlOjbects("task", b)

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
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

func TestTomlTasks(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		b := []byte{}
		got, err := tomlTasks(b, defaultConfig{})

		if err != nil {
			t.Errorf("got: %v, want: %v", err, nil)
		}

		if len(got) != 0 {
			t.Errorf("got: %d, want: %d", len(got), 0)
		}
	})

	t.Run("simple task", func(t *testing.T) {
		b, err := os.ReadFile("./examples/task_simple.toml")
		if err != nil {
			panic(err)
		}

		d, _ := tomlDefaultConfig(extractTomlOjbects("default", b)[0])
		got, err := tomlTasks(b, d)

		if err != nil {
			t.Errorf("got: %s, want: %v", err, nil)
		}

		if len(got) != 1 {
			t.Errorf("got: %d, want: %d", len(got), 1)
		}

		if got[0].Name != "simple task" {
			t.Errorf("got: %s, want: %s", got[0].Name, "simple task")
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
}

func TestMergeEnvVars(t *testing.T) {
	var tests = []struct {
		a    []string
		b    []string
		want []string
	}{
		{ // 0
			[]string{"A=1"},
			[]string{"B=2"},
			[]string{"A=1", "B=2"},
		},
		{ // 1
			[]string{"A=1"},
			[]string{"A=2"},
			[]string{"A=1"},
		},
		{ // 1
			[]string{"A=1"},
			[]string{"B=2", "C=3"},
			[]string{"A=1", "B=2", "C=3"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got := mergeEnvVars(tt.a, tt.b)
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
