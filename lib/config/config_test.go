package config

import (
	"fmt"
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	fhA, _ := os.Open("./examples/A.toml")

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
				t.Errorf("default timeout got: %d, want: %d", def.Retries, tt.wantDef.Retries)
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

			// unset the global state
			def = defaultConfig{}
			Directory = ""
		})
	}
}

func TestExtractTomlObjects(t *testing.T) {
	f, _ := os.Open("./examples/A.toml")
	b, err := PreProcess(f, []*os.File{})
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	t.Run("task", func(t *testing.T) {
		got := extractTomlOjbects("task", b)

		if len(got) != 1 {
			t.Errorf(" got: %d, want: %d", len(got), 1)
		}
	})

	t.Run("gate", func(t *testing.T) {
		got := extractTomlOjbects("gate", b)

		if len(got) != 1 {
			t.Errorf(" got: %d, want: %d", len(got), 1)
		}
	})

	t.Run("notifier", func(t *testing.T) {
		got := extractTomlOjbects("notifier", b)

		if len(got) != 1 {
			t.Errorf(" got: %d, want: %d", len(got), 1)
		}
	})
}
