package fz

import (
	"fmt"
	"testing"
)

func TestMergeEnvVars(t *testing.T) {
	var tests = []struct {
		a    []string
		b    []string
		want []string
	}{
		{ // 0
			[]string{"A=1"},
			[]string{"B=2"},
			[]string{"A=1", "B=2"},
		},
		{ // 1
			[]string{"A=1"},
			[]string{"A=2"},
			[]string{"A=1"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got := MergeEnvVars(tt.a, tt.b)
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
