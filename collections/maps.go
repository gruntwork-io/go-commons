package collections

import (
	"fmt"
	"sort"
)

const (
	DefaultKeyValueStringSliceFormat = "%s=%s"
)

// Merge all the maps into one. Sadly, Go has no generics, so this is only defined for string to interface maps.
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}

	for _, currMap := range maps {
		for key, value := range currMap {
			out[key] = value
		}
	}

	return out
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
