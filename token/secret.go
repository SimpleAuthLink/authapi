package token

import "crypto/sha256"

type Secret []byte

func (s *Secret) SetParts(raw ...[]byte) *Secret {
	if s == nil {
		s = new(Secret)
	}
	newParts := []byte{}
	for _, part := range raw {
		if len(part) == 0 {
			continue
		}
		hsecret := sha256.Sum256(part)
		newParts = append(newParts, hsecret[:]...)
	}
	if len(newParts) != 0 {
		*s = append(*s, newParts...)
	}
	return s
}

func (s *Secret) Bytes() []byte {
	return []byte(*s)
}

func (s *Secret) Valid() bool {
	if s == nil {
		return false
	}
	// secret is valid if it has more than 1 part, and each part is hashed
	// to a sha256 size
	return len(*s) > sha256.Size
}
