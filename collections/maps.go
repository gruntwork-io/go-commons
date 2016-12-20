package collections

// Merge all the maps into one. Sadly, Go has no generics, so this is only defined for string to interface maps.
func MergeMaps(maps ... map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}

	for _, currMap := range maps {
		for key, value := range currMap {
			out[key] = value
		}
	}

	return out
}

