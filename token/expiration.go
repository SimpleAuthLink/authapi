package token

import (
	"encoding/base64"
	"time"
)

type Expiration time.Time

func NewExpiration(d time.Duration) Expiration {
	return Expiration(time.Now().Add(d))
}

func (exp *Expiration) String() string {
	t := time.Time(*exp)
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func (exp *Expiration) SetString(data string) *Expiration {
	t, err := time.Parse(time.RFC3339Nano, data)
	if err != nil {
		return nil
	}
	*exp = Expiration(t)
	return exp
}

func (exp *Expiration) Bytes() []byte {
	if exp.String() == "" {
		return nil
	}
	return []byte(exp.String())
}

func (exp *Expiration) SetBytes(data []byte) *Expiration {
	return exp.SetString(string(data))
}

func (exp *Expiration) Marshal() []byte {
	bExp := exp.Bytes()
	if len(bExp) == 0 || bExp[0] == 0 {
		return nil
	}
	b := make([]byte, base64.RawStdEncoding.EncodedLen(len(bExp)))
	base64.RawStdEncoding.Encode(b, bExp)
	return b
}

func (exp *Expiration) Unmarshal(data []byte) *Expiration {
	b := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
	if _, err := base64.RawStdEncoding.Decode(b, data); err != nil {
		return nil
	}
	return exp.SetBytes(b)
}
