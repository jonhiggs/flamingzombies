package trace

import (
	"fmt"
	"regexp"
	"testing"
)

func TestTraceID(t *testing.T) {
	got := New()

	if len(fmt.Sprint(got)) != 16 {
		t.Errorf("%v: should have length of 16", got)
	}

	re := regexp.MustCompile(`^[a-z0-9]*$`)
	if !re.Match([]byte(fmt.Sprintf("%s", got))) {
		t.Errorf("%v: must contain only numbers and lowercase letters", got)
	}
}
