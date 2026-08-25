package token

import (
	"bytes"
	"crypto/sha256"
)

// secretHashSize is the size of the secret hash. It is used to determine
// the size of the secret when it is hashed. The hash is created by hashing
// the secret to a sha256 size, but then it is truncated to 12 bytes.
// The hash is used to sign and verify tokens, and to create the app ID and it is part of it.
const secretHashSize = 12

// minSecretParts is the minimum number of parts that a valid secret should
// have. The minimun number of required parts is 2 to ensure that the ownership
// of the secret is distributed into at least two parties.
const minSecretParts = 2

// Secret represents a secret that is used to sign and verify tokens. It is
// a wrapper around a byte slice that provides additional methods for setting
// and getting the secret. It should have at least 2 parts, each hashed to a
// sha256 size.
type Secret []byte

// SetParts method sets the secret's parts from a slice of byte slices. If the
// secret is nil, a new secret is created. If the parts are empty, the secret
// is not set. The parts are hashed to a sha256 size and concatenated to form
// the secret.
func (s *Secret) SetParts(raw ...[]byte) *Secret {
	// if no secret is provided, initialize a new one
	if s == nil {
		s = new(Secret)
	}
	// hash each part to a sha256 size and concatenate them
	newParts := []byte{}
	for _, rawPart := range raw {
		part := bytes.TrimSpace(rawPart)
		if len(part) == 0 {
			continue
		}
		hsecret := sha256.Sum256(part)
		newParts = append(newParts, hsecret[:]...)
	}
	// if there are new parts, append them to the secret
	if len(newParts) > 0 {
		*s = append(*s, newParts...)
	}
	return s
}

// Bytes method returns the secret as a byte slice.
func (s *Secret) Bytes() []byte {
	return []byte(*s)
}

// Hash method returns the hash of the secret as a byte slice. The hash is
// created by hashing the secret to a sha256 size. The hash is used to create
// the app ID and it is part of it, but is also used to sign and verify the
// user sessions in the token generation process.
func (s *Secret) Hash() []byte {
	if s == nil {
		return nil
	}
	// hash the secret to a sha256 size
	h := sha256.Sum256(*s)
	return h[:secretHashSize]
}

// Valid method returns true if the secret is valid, false otherwise. A secret
// is considered valid if it has more than 1 part, and each part is hashed to
// a sha256 size.
func (s *Secret) Valid() bool {
	if s == nil {
		return false
	}
	// secret is valid if it has more than 1 part, and each part is hashed
	// to a sha256 size
	lenSecret := len(*s)
	validSize := lenSecret%sha256.Size == 0
	validNumOfParts := lenSecret/sha256.Size >= minSecretParts
	return validSize && validNumOfParts
}
