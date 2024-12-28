package fz

import (
	"strings"
)

// Merge and deduplicate environment variables
func MergeEnvVars(a, b []string) []string {
	var res, keys []string
	for _, s := range a {
		found := false
		k := strings.Split(s, "=")[0]

		for _, ks := range keys {
			if k == ks {
				found = true
			}
		}

		if !found {
			keys = append(keys, k)
			res = append(res, s)
		}
	}

	for _, s := range b {
		found := false
		k := strings.Split(s, "=")[0]

		for _, ks := range keys {
			if k == ks {
				found = true
			}
		}

		if !found {
			keys = append(keys, k)
			res = append(res, s)
		}
	}

	return res
}
