package random

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomStringIsMostlyRandom(t *testing.T) {
	t.Parallel()

	// Ensure that there is no overlap in 32 character random strings generated 100 times
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		newStr, err := RandomString(32, Base62Chars)
		require.NoError(t, err)
		_, hasSeen := seen[newStr]
		require.False(t, hasSeen)
		seen[newStr] = true
	}
}

func TestRandomStringRespectsStrLen(t *testing.T) {
	t.Parallel()

	for i := 0; i < 40; i++ {
		newStr, err := RandomString(i, Base62Chars)
		require.NoError(t, err)
		assert.Equal(t, len(newStr), i)
	}
}

func TestRandomStringRespectsAllowedChars(t *testing.T) {
	t.Parallel()

	for i := 0; i < 100; i++ {
		newStr, err := RandomString(10, Digits)
		require.NoError(t, err)
		// Since the new string should only be composed of digits, if RandomString respects allowed chars you should
		// always be able to convert the string to an int
		_, err = strconv.Atoi(newStr)
		require.NoError(t, err)
	}
}
