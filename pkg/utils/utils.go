package utils

import (
	"sort"
)

func SortedKeys(in map[string]int) []string {
	out := make([]string, 0, len(in))
	for key := range in {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
