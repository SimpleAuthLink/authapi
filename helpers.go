package authapi

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
)

// randBytes generates a random byte slice of length n. It returns nil if n is
// less than 1.
func randBytes(n int) []byte {
	if n < 1 {
		return nil
	}
	b := make([]byte, n)
	for i := 0; i < n; {
		val := rand.Uint64()
		for j := 0; j < 8 && i < n; j++ {
			b[i] = byte(val & 0xff)
			val >>= 8
			i++
		}
	}
	return b
}

// hash generates a hash of the input string using SHA-256 algorithm. The n
// parameter allows to truncate the hash to n bytes. It returns the hash as a
// hexadecimal string. The resulting string will have a length of 2*n. If n is
// less than 1 or greater than the hash length, the full hash will be returned.
// If the input string is empty, it returns an empty string. If something fails
// during the hashing process, it returns an error.
func hash(input string, n int) (string, error) {
	if input == "" {
		return "", nil
	}
	hash := sha256.New()
	if _, err := hash.Write([]byte(input)); err != nil {
		return "", err
	}
	bHash := hash.Sum(nil)
	if n > 0 && n < len(bHash) {
		bHash = bHash[:n]
	}
	return hex.EncodeToString(bHash), nil
}
