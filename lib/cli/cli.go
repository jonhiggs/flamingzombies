package cli

import (
	"fmt"
	"os"
)

func Error(m interface{}) {
	msg := fmt.Sprintf("FATAL: %v", m)
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
