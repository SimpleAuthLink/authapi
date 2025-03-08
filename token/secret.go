package token

import "crypto/sha256"

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
	for _, part := range raw {
		if len(part) == 0 {
			continue
		}
		hsecret := sha256.Sum256(part)
		newParts = append(newParts, hsecret[:]...)
	}
	// if there are new parts, append them to the secret
	if len(newParts) != 0 {
		*s = append(*s, newParts...)
	}
	return s
}

// Bytes method returns the secret as a byte slice.
func (s *Secret) Bytes() []byte {
	return []byte(*s)
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
	return len(*s) > sha256.Size
}
