// Heavily refer to this:
// https://www.calhoun.io/creating-random-strings-in-go/

package randomid

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// StringNWithCharset generate a random string according on charset
func StringNWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// StringN ...
func StringN(length int) string {
	return StringNWithCharset(length, charset)
}
