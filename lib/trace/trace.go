package trace

import (
	"crypto/rand"
	"fmt"
	"strings"
)

type TraceID struct {
	Bytes []byte
}

func New() TraceID {
	var id TraceID
	id.Bytes = make([]byte, 8)

	_, err := rand.Read(id.Bytes)
	if err != nil {
		fmt.Println("Error: ", err)
		return TraceID{}
	}

	return id
}

func (id TraceID) String() string {
	return strings.ToLower(fmt.Sprintf("%X", id.Bytes))
}
