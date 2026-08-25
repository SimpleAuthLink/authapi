package token

import (
	"bytes"
	"net/mail"

	"go.k7z7z.cc/x/encoding/base64url"
)

// Email type wraps a byte slice to handle an user email regarding the token
// generation, managment and validation.
type Email []byte

// Valid method returns a boolean indicating if the current email is a valid
// email address or not. It returns false if the email is nil, empty or invalid
// address.
func (e *Email) Valid() bool {
	// If the current email is nil, return false
	if e == nil {
		return false
	}
	// Trim the given email
	value := bytes.TrimSpace(e.Bytes())
	if len(value) == 0 {
		return false
	}
	// Ensure that the trimmed email is equal to the original one after parse
	// it as an email address.
	addr, err := mail.ParseAddress(string(value))
	if err != nil {
		return false
	}
	return addr.Address == string(value)
}

// SetBytes method writes a byte slice to the current Email type instance. If
// a nil email is used, it is initialized.
func (e *Email) SetBytes(bEmail []byte) *Email {
	if e == nil {
		e = new(Email)
	}
	*e = bEmail
	return e
}

// Bytes method return the byte slice value of the current Email type.
func (e *Email) Bytes() []byte {
	return *e
}

// Marshal method encodes the current email using a url safe version of base64
// encoding.
func (e *Email) Marshal() []byte {
	if e == nil {
		return nil
	}
	return base64url.RawEncode(*e)
}

// Unmarshal method decodes the base64 encoded email into a Email type
// instance. If the current email is not initialized, it is initialized
// first. If decoding fails, it returns the current Email instance instead.
func (e *Email) Unmarshal(raw []byte) *Email {
	if e == nil {
		e = new(Email)
	}
	dec, err := base64url.RawDecode(raw)
	if err != nil {
		return e
	}
	return e.SetBytes(dec)
}

// SetString method rites the stirng input as byte slice to the current Email
// type instance. It wraps the SetBytes method.
func (e *Email) SetString(email string) *Email {
	return e.SetBytes([]byte(email))
}

// String method return the string value of the current Email type.
func (e *Email) String() string {
	return string(e.Bytes())
}
