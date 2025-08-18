package token

import (
	"time"

	"github.com/simpleauthlink/authapi/internal/base64url"
)

// Expiration represents a time when a token expires. It is a wrapper around
// time.Time that provides additional methods for setting and getting the
// expiration time.
type Expiration time.Time

// Valid method returns true if the expiration is valid, false otherwise. An
// expiration is considered valid if it is in the future.
func (exp *Expiration) Valid() bool {
	return time.Now().Before(exp.Time())
}

// Time method returns the expiration time as a time.Time.
func (exp *Expiration) Time() time.Time {
	if exp == nil {
		return time.Time{}
	}
	return time.Time(*exp)
}

// SetTime method sets the expiration time from a time.Time. If the expiration
// is nil, a new expiration is created.
func (exp *Expiration) SetTime(t time.Time) *Expiration {
	// if no expiration is provided, initialize a new one
	if exp == nil {
		exp = new(Expiration)
	}
	// set the expiration time
	*exp = Expiration(t)
	// if the expiration is invalid, return nil
	if !exp.Valid() {
		return nil
	}
	return exp
}

// Duration method returns the duration until the expiration time.
func (exp *Expiration) Duration() time.Duration {
	return time.Until(exp.Time())
}

// SetDuration method sets the expiration time from a duration. If the duration
// is invalid, the expiration time is not set.
func (exp *Expiration) SetDuration(d time.Duration) *Expiration {
	if d < minDuration || d > maxDuration {
		return nil
	}
	if exp == nil {
		exp = new(Expiration)
	}
	return exp.SetTime(time.Now().Add(d))
}

// String method returns the expiration time as a string in RFC3339Nano format.
// It is useful for encoding the expiration time. If the expiration is nil, an
// empty string is returned.
func (exp *Expiration) String() string {
	if exp == nil {
		return ""
	}
	if !exp.Valid() {
		return ""
	}
	return exp.Time().Format(time.RFC3339Nano)
}

// SetString method sets the expiration time from a string in RFC3339Nano
// format. It is useful for decoding the expiration time. If the string
// is invalid, the expiration time is not set and nil is returned. If the
// expiration is nil, a new expiration is created. If the resulting expiration
// is invalid, nil is returned.
func (exp *Expiration) SetString(data string) *Expiration {
	// parse the expiration time
	t, err := time.Parse(time.RFC3339Nano, data)
	if err != nil {
		return nil
	}
	// if the expiration is nil, initialize a new one
	if exp == nil {
		exp = new(Expiration)
	}
	// set the expiration time
	*exp = Expiration(t)
	// if the expiration is invalid, return nil
	if !exp.Valid() {
		return nil
	}
	return exp
}

// Bytes method returns the expiration time as a byte slice. It is useful for
// encoding the expiration time. If the expiration is nil, nil is returned. It
// is equivalent to converting the expiration time to a string and then
// converting the string to a byte slice.
func (exp *Expiration) Bytes() []byte {
	if exp.String() == "" {
		return nil
	}
	return []byte(exp.String())
}

// SetBytes method sets the expiration time from a byte slice. It is useful for
// decoding the expiration time. It is equivalent to converting the byte slice
// to a string and then setting the expiration time from the string.
func (exp *Expiration) SetBytes(data []byte) *Expiration {
	return exp.SetString(string(data))
}

// Marshal method returns the expiration time as a base64 encoded byte slice. It
// is useful for encoding the expiration time. If the expiration is nil or
// invalid, nil is returned.
func (exp *Expiration) Marshal() []byte {
	bExp := exp.Bytes()
	if len(bExp) == 0 || bExp[0] == 0 {
		return nil
	}
	return base64url.RawEncode(bExp)
}

// Unmarshal method sets the expiration time from a base64 encoded byte slice. It
// is useful for decoding the expiration time. If the expiration is nil or
// invalid, nil is returned.
func (exp *Expiration) Unmarshal(data []byte) *Expiration {
	b, err := base64url.RawDecode(data)
	if err != nil {
		return nil
	}
	return exp.SetBytes(b)
}
