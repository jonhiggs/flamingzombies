package trace

import (
	"regexp"
	"testing"
)

func TestTraceID(t *testing.T) {
	got := ID()

	if len(got) != 16 {
		t.Errorf("%v: should have length of 16", got)
	}

	re := regexp.MustCompile(`^[a-z0-9]*$`)
	if !re.Match([]byte(got)) {
		t.Errorf("%v: must contain only numbers and lowercase letters", got)
	}
}
