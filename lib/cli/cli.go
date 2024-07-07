package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func Error(m interface{}) {
	msg := fmt.Sprintf("Error: %v", m)
	fmt.Fprintln(os.Stderr, msg)
	// publish an error metric
	os.Exit(3)
}

func Timeout() {
	fmt.Fprintln(os.Stderr, "Timeout: task took too long to execute")
	// publish a timeout metric
	os.Exit(3)
}

// get the timeout from the environment, or use the default.
func GetTimeout(v, def, min time.Duration) time.Duration {
	if v != 0 {
		return v
	}

	if os.Getenv("TIMEOUT") == "" {
		return def
	}

	t, err := strconv.Atoi(os.Getenv("TIMEOUT"))
	if err != nil {
		Error(err)
	}

	v = time.Duration(t) * time.Second

	if v < min {
		Error(fmt.Sprintf("Timeout is less than minimum of %v"))
	}

	return v
}

func Version(app, version string) {
	fmt.Printf("%s %s\n", app, version)
	os.Exit(0)
}
