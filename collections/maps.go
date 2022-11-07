package collections

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

const (
	DefaultKeyValueStringSliceFormat = "%s=%s"
)

// MergeMaps merges all the maps into one
func MergeMaps[K comparable, V any](mapsToMerge ...map[K]V) map[K]V {
	result := map[K]V{}

	for _, currMap := range mapsToMerge {
		maps.Copy(result, currMap)
	}

	return result
}

// Return the keys for the given map, sorted alphabetically
func Keys(m map[string]string) []string {
	out := []string{}

	for key, _ := range m {
		out = append(out, key)
	}

	sort.Strings(out)

	return out
}

// KeyValueStringSlice returns a string slice with key=value items, sorted alphabetically
func KeyValueStringSlice(m map[string]string) []string {
	return KeyValueStringSliceWithFormat(m, DefaultKeyValueStringSliceFormat)
}

// KeyValueStringSliceWithFormat returns a string slice using the specified format, sorted alphabetically.
// The format should consist of at least two '%s' string verbs.
func KeyValueStringSliceWithFormat(m map[string]string, format string) []string {
	out := []string{}

	for key, value := range m {
		out = append(out, fmt.Sprintf(format, key, value))
	}

	sort.Strings(out)

	return out
}

// KeyValueStringSliceAsMap converts a string slice with key=value items into a map of slice values. The slice will
// contain more than one item if a key is repeated in the string slice list.
func KeyValueStringSliceAsMap(kvPairs []string) map[string][]string {
	out := make(map[string][]string)
	for _, kvPair := range kvPairs {
		x := strings.Split(kvPair, "=")
		key := x[0]

		var value string
		if len(x) > 1 {
			value = strings.Join(x[1:], "=")
		}

		if _, hasKey := out[key]; hasKey {
			out[key] = append(out[key], value)
		} else {
			out[key] = []string{value}
		}
	}
	return out
}
