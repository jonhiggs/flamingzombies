package trace

import (
	"crypto/rand"
	"fmt"
	"strings"
)

func ID() string {
	b := make([]byte, 8)

	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return ""
	}

	return strings.ToLower(fmt.Sprintf("%X", b))
}
