package trace

import (
	"crypto/rand"
	"fmt"
	"strings"
)

type ID struct {
	Bytes []byte
}

func New() ID {
	var id ID
	id.Bytes = make([]byte, 8)

	_, err := rand.Read(id.Bytes)
	if err != nil {
		fmt.Println("Error: ", err)
		return ID{}
	}

	return id
}

func (id ID) String() string {
	return strings.ToLower(fmt.Sprintf("%X", id.Bytes))
}
