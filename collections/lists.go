package collections

// Return true if the given list contains the given element
func ListContainsElement(list []string, element string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}

	return false
}

// Return a copy of the given list with all instances of the given element removed
func RemoveElementFromList(list []string, element string) []string {
	out := []string{}
	for _, item := range list {
		if item != element {
			out = append(out, item)
		}
	}
	return out
}

// BatchListIntoGroupsOf will group the provided string slice into groups of size n, with the last of being truncated to
// the remaining count of strings.  Returns nil if n is <= 0
func BatchListIntoGroupsOf(slice []string, batchSize int) [][]string {
	if batchSize <= 0 {
		return nil
	}

	// We make a copy of the slice here so that we can modify it
	copyOfSlice := make([]string, len(slice))
	copy(copyOfSlice, slice)

	// Taken from SliceTricks: https://github.com/golang/go/wiki/SliceTricks#batching-with-minimal-allocation
	// Intuition: We repeatedly slice off batchSize elements from copyOfSlice and append it to the output, until there
	// is not enough.
	output := [][]string{}
	for batchSize < len(copyOfSlice) {
		copyOfSlice, output = copyOfSlice[batchSize:], append(output, copyOfSlice[0:batchSize:batchSize])
	}
	if len(copyOfSlice) > 0 {
		output = append(output, copyOfSlice)
	}
	return output
}
