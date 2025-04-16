package token

import "bytes"

// Token is a type that represents a user token. It is a wrapper around a byte
// slice that provides additional methods for setting and getting the token. It
// should have 2 parts, the first part is the expiration time, and the second
// part is the signature.
type Token []byte

// String method returns the token as a string. It is useful for encoding the
// token. If the token is nil, an empty string is returned. It internally calls
// the Bytes method to get the token as a byte slice.
func (t *Token) String() string {
	if t == nil {
		return ""
	}
	return string(t.Bytes())
}

// SetString method sets the token from a string. It is useful for decoding the
// token. The string should be the token's expiration time and signature joined
// by the token separator. If the token is invalid, the token is not set. It
// internally calls the SetBytes method to set the token from a byte slice.
func (t *Token) SetString(data string) *Token {
	if t == nil {
		t = new(Token)
	}
	return t.SetBytes([]byte(data))
}

// Bytes method returns the token as a byte slice. It is useful for encoding
// the token. If the token is nil, nil is returned. It internally calls the
// parts method to get the token's expiration time and signature as byte
// slices. It checks that the parts are valid before returning the token
// as a byte slice.
func (t *Token) Bytes() []byte {
	// check if the token is nil
	if t == nil {
		return nil
	}
	// check if the token has valid parts
	if _, _, ok := t.parts(); !ok {
		return nil
	}
	// return the token as a byte slice
	return []byte(*t)
}

// SetBytes method sets the token from a byte slice. It is useful for
// decoding the token. The byte slice should be the token's expiration time
// and signature joined by the token separator. If the token is invalid, the
// token is not set. It internally calls the parts method to get the token's
// expiration time and signature as byte slices. It checks that the parts are
// valid before setting the token from the byte slice.
func (t *Token) SetBytes(data []byte) *Token {
	// if no token is provided, create a new one
	if t == nil {
		t = new(Token)
	}
	// generate a new token from the data
	ntoken := &Token{}
	*ntoken = data
	// check if the new token has valid parts
	if _, _, ok := ntoken.parts(); !ok {
		return t
	}
	// set the token to the new token
	*t = data
	return t
}

// Expiration method returns the token's expiration time. It is useful for
// getting the expiration time. If the token is nil, nil is returned. It
// internally calls the parts method to get the token's expiration time and
// signature as byte slices. It checks that the expiration time is valid before
// returning it.
func (t *Token) Expiration() *Expiration {
	if rawExp, _, ok := t.parts(); ok {
		return new(Expiration).Unmarshal(rawExp)
	}
	return nil
}

// SetExpiration method sets the token's expiration time. If the token is nil,
// a new token is created. If the expiration time is invalid, the token is not
// set. It internally calls the parts method to replace the token's expiration
// time with the new expiration time. It checks that the expiration time is
// valid before setting it.
func (t *Token) SetExpiration(exp Expiration) *Token {
	// if no token is provided, create a new one
	if t == nil {
		t = new(Token)
	}
	// if expiration is invalid, return the current token
	if !exp.Valid() {
		return t
	}
	// create a base content with the new expiration and no signature
	baseContent := append(exp.Marshal(), tokenSeparator)
	// get the current signature, if there is one, return a token with the
	// base content and no signature
	sig := t.Signature()
	if sig == nil {
		return t.SetBytes(baseContent)
	}
	// if there is a signature, update the token to replace the expiration
	// part with the new expiration
	return t.SetBytes(append(baseContent, sig...))
}

// Signature method returns the token's signature. If the token is nil, nil is
// returned. It internally calls the parts method to get the signature part as
// byte slices.
func (t *Token) Signature() []byte {
	_, sig, _ := t.parts()
	return sig
}

// SetSignature method sets the token's signature. If the token is nil, a
// new token is created. If the signature is nil, the token is not set. It
// internally calls the parts method to replace the token's signature with
// the new signature.
func (t *Token) SetSignature(sig []byte) *Token {
	// if no token is provided, create a new one
	if t == nil {
		t = new(Token)
	}
	// create a base content with no expiration and the new signature
	baseContent := append([]byte{tokenSeparator}, sig...)
	// get the current expiration, if there is none, return a token with the
	// base content and no expiration
	exp := t.Expiration()
	if exp == nil {
		return t.SetBytes(baseContent)
	}
	// if there is an expiration, update the token to replace the signature
	// part with the new signature
	return t.SetBytes(append(exp.Marshal(), baseContent...))
}

// parts private method returns the token's expiration time and signature as
// byte slices. It also returns a boolean indicating if the token has valid
// parts. If the token is nil, or the parts are invalid, the parts are nil and
// the boolean is false. It splits the token by the token separator and checks
// that the result has 2 parts (expiration time and signature).
func (t *Token) parts() ([]byte, []byte, bool) {
	if t == nil {
		return nil, nil, false
	}
	if p := bytes.Split([]byte(*t), []byte{tokenSeparator}); len(p) == 2 {
		return p[0], p[1], true
	}
	return nil, nil, false
}
