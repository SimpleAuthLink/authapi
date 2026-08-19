package token

import "go.k7z7z.cc/x/encoding/base64url"

type Email []byte

func (e *Email) SetBytes(bEmail []byte) *Email {
	if e == nil {
		e = new(Email)
	}
	*e = bEmail
	return e
}

func (e *Email) Bytes() []byte {
	return *e
}

func (e *Email) Marshal() []byte {
	if e == nil {
		return nil
	}
	return base64url.RawEncode(*e)
}

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

func (e *Email) SetString(email string) *Email {
	return e.SetBytes([]byte(email))
}

func (e *Email) String() string {
	return string(e.Bytes())
}
