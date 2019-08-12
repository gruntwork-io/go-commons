// Package random provides utilities and functions for generating random data.
package random

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

// Character sets that you can use when passing into RandomString
const Digits = "0123456789"
const UpperLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const LowerLetters = "abcdefghijklmnopqrstuvwxyz"
const SpecialChars = "<>[]{}()-_*%&/?\"'\\"

var Base62Chars = Digits + UpperLetters + LowerLetters

// RandomString generates a random string of length strLength, composing only of characters in allowedChars. Based on
// code here: http://stackoverflow.com/a/9543797/483528
func RandomString(strLength int, allowedChars string) (string, error) {
	var out bytes.Buffer

	for i := 0; i < strLength; i++ {
		id, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedChars))))
		if err != nil {
			return out.String(), err
		}
		out.WriteByte(allowedChars[id.Int64()])
	}

	return out.String(), nil
}
