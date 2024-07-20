package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
)

// EncodeUserToken function encodes the user information into a token and
// returns it. It receives the app id and the email of the user and returns the
// token and the user id. If the app id or the email are empty, it returns an
// error. The token is composed of three parts separated by a token separator.
// The first part is a random sequence of 8 bytes encoded as a hexadecimal
// string. The second part is the app id and the third part is the user id. The
// user id is generated hashing the email with a length of 4 bytes. The token
// is returned following the token format:
//
//	[appId(8)]-[userId(8)]-[randomPart(16)]
func EncodeUserToken(appId, email string) (string, string, error) {
	// check if the app id and email are not empty
	if len(appId) == 0 || len(email) == 0 {
		return "", "", fmt.Errorf("appId and email are required")
	}
	bToken := RandBytes(TokenSize)
	hexToken := hex.EncodeToString(bToken)
	// hash email
	userId, err := Hash(email, UserIdSize)
	if err != nil {
		return "", "", err
	}
	return strings.Join([]string{appId, userId, hexToken}, TokenSeparator), userId, nil
}

// DecodeUserToken function decodes the user information from the token provided
// and returns the app id and the user id. If the token is invalid, it returns
// an error. It splits the provided token by the token separator and returns the
// second and third parts, which are the app id and the user id respectively,
// following the token format:
//
//	[appId(8)]-[userId(8)]-[randomPart(16)]
func DecodeUserToken(token string) (string, string, error) {
	tokenParts := strings.Split(token, TokenSeparator)
	if len(tokenParts) != 3 {
		return "", "", fmt.Errorf("invalid token")
	}
	return tokenParts[0], tokenParts[1], nil
}

// RandBytes generates a random byte slice of length n. It returns nil if n is
// less than 1.
func RandBytes(n int) []byte {
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

// Hash generates a hash of the input string using SHA-256 algorithm. The n
// parameter allows to truncate the hash to n bytes. It returns the hash as a
// hexadecimal string. The resulting string will have a length of 2*n. If n is
// less than 1 or greater than the hash length, the full hash will be returned.
// If the input string is empty, it returns an empty string. If something fails
// during the hashing process, it returns an error.
func Hash(input string, n int) (string, error) {
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
