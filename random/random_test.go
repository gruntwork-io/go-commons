package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomStringIsMostlyRandom(t *testing.T) {
	t.Parallel()

	// Ensure that there is no overlap in 32 character random strings generated 100 times
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		newStr := RandomString(32, Base62Chars)
		_, hasSeen := seen[newStr]
		require.False(t, hasSeen)
		seen[newStr] = true
	}
}

func TestRandomStringRespectsStrLen(t *testing.T) {
	t.Parallel()

	for i := 0; i < 40; i++ {
		newStr := RandomString(i, Base62Chars)
		assert.Equal(t, len(newStr), i)
	}
}
