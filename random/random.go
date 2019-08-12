// Package random provides utilities and functions for generating random data.
package random

import (
	"bytes"
	"math/rand"
	"time"
)

// Character sets that you can use when passing into RandomString
const Digits = "0123456789"
const UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const LowerLetters = "abcdefghijklmnopqrstuvwxyz"
const SpecialChars = "<>[]{}()-_*%&/?\"'\\"

var Base62Chars = Digits + UpperLetters + LowerLetters

// RandomString generates a random string of length strLength, composing only of characters in allowedChars. Based on
// code here: http://stackoverflow.com/a/9543797/483528
func RandomString(strLength int, allowedChars string) string {
	var out bytes.Buffer

	generator := newRand()
	for i := 0; i < strLength; i++ {
		out.WriteByte(allowedChars[generator.Intn(len(allowedChars))])
	}

	return out.String()
}

// newRand creates a new random number generator, seeding it with the current system time.
func newRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}
