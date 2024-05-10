package authapi

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
)

func randBytes(n int) []byte {
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

func hash(input string, n int) (string, error) {
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
