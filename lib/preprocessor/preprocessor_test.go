package preprocessor

import (
	"fmt"
	"os"
	"testing"
)

func TestExampleA(t *testing.T) {
	fhA, _ := os.Open("./examples/A.toml")
	fhB, _ := os.Open("./examples/B.toml")
	fhC, _ := os.Open("./examples/C.toml")
	fhD, _ := os.Open("./examples/D.toml")
	fhE, _ := os.Open("./examples/E.toml")
	fhF, _ := os.Open("./examples/F.toml")

	var tests = []struct {
		fh      *os.File
		want    []byte
		wantErr error
	}{
		{ // A
			fhA,
			[]byte(
				`# A
log_file = "-"
log_level = "info"

[defaults]
retries = 5
timeout = 1
notifiers = []
`),
			nil,
		},
		{ // B
			fhB,
			[]byte(
				`# B
[[task]]
name = "example_b"
command = "task/example"
frequency = 20
`),
			nil,
		},
		{ // C
			fhC,
			[]byte(
				`# A
log_file = "-"
log_level = "info"

[defaults]
retries = 5
timeout = 1
notifiers = []
`),
			nil,
		},
		{ // D
			fhD,
			[]byte(
				`# D
# A
log_file = "-"
log_level = "info"

[defaults]
retries = 5
timeout = 1
notifiers = []

# B
[[task]]
name = "example_b"
command = "task/example"
frequency = 20
`),
			nil,
		},
		{ // E
			fhE,
			[]byte(
				`# E
# B
[[task]]
name = "example_b"
command = "task/example"
frequency = 20
# B
[[task]]
name = "example_b"
command = "task/example"
frequency = 20
`),
			nil,
		},
		{ // F
			fhF,
			[]byte{},
			ErrCircularDependency,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.fh.Name()), func(t *testing.T) {
			got, err := Run(tt.fh, []*os.File{})
			if string(got) != string(tt.want) {
				t.Errorf("\ngot: %s\nwant: %s\n", got, tt.want)
			}

			if err != tt.wantErr {
				t.Errorf("\ngot: %s\nwant: %s\n", err, tt.wantErr)
			}
		})
	}
}
