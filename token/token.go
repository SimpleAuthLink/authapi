package token

import (
	"bytes"
)

type Token []byte

func (t *Token) String() string {
	if t == nil {
		return ""
	}
	return string(t.Bytes())
}

func (t *Token) SetString(data string) *Token {
	if t == nil {
		t = new(Token)
	}
	return t.SetBytes([]byte(data))
}

func (t *Token) Bytes() []byte {
	if t == nil {
		return nil
	}
	if _, _, ok := t.parts(); !ok {
		return nil
	}
	return []byte(*t)
}

func (t *Token) SetBytes(data []byte) *Token {
	if t == nil {
		t = new(Token)
	}
	ntoken := &Token{}
	*ntoken = data
	if _, _, ok := ntoken.parts(); !ok {
		return t
	}
	*t = data
	return t
}

func (t *Token) Expiration() *Expiration {
	if rawExp, _, ok := t.parts(); ok {
		return new(Expiration).Unmarshal(rawExp)
	}
	return nil
}

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

func (t *Token) Signature() []byte {
	_, sig, _ := t.parts()
	return sig
}

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

func (t *Token) parts() ([]byte, []byte, bool) {
	if t == nil {
		return nil, nil, false
	}
	if p := bytes.Split([]byte(*t), []byte{tokenSeparator}); len(p) == 2 {
		return p[0], p[1], true
	}
	return nil, nil, false
}
