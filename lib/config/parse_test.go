package config

import (
	"fmt"
	"os"
	"testing"
)

func TestParseDefault(t *testing.T) {
	fhA, _ := os.Open("./examples/A.toml")

	var tests = []struct {
		fh      *os.File
		want    defaultConfig
		wantErr error
	}{
		{ // A
			fhA,
			defaultConfig{
				Retries:        5,
				TimeoutSeconds: 1,
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.fh.Name()), func(t *testing.T) {
			b, err := PreProcess(tt.fh, []*os.File{})
			if err != nil {
				panic(fmt.Errorf("while preprocessing: %w", err))
			}

			got, err := parseDefault(b)

			if got.Retries != tt.want.Retries {
				t.Errorf("retries got: %d,want: %d", got.Retries, tt.want.Retries)
			}

			if got.TimeoutSeconds != tt.want.TimeoutSeconds {
				t.Errorf("retries got: %d,want: %d", got.Retries, tt.want.Retries)
			}

			if err != tt.wantErr {
				t.Errorf("\ngot: %s\nwant: %s\n", err, tt.wantErr)
			}
		})
	}
}
