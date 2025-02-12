package token

import (
	"bytes"
	"testing"
	"time"
)

func TestNewExpirationTime(t *testing.T) {
	exp := NewExpiration(minDuration - 1)
	if exp != nil {
		t.Fatalf("expected nil, got %v", exp)
	}
	exp = NewExpiration(minDuration * 2)
	if exp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	expTime := exp.Time()
	expected := time.Now().Add(minDuration * 2)
	if expected.Sub(expTime) > time.Millisecond*50 {
		t.Errorf("expected %v, got %v", expected, expTime)
	}
}

func TestExpirationValid(t *testing.T) {
	t.Parallel()
	exp := NewExpiration(minDuration)
	if !exp.Valid() {
		t.Errorf("expected valid expiration, got invalid")
	}
	time.Sleep(minDuration)
	if exp.Valid() {
		t.Errorf("expected invalid expiration, got valid")
	}
}

func TestStringSetStringExpiration(t *testing.T) {
	exp := NewExpiration(minDuration)
	str := exp.String()
	decoded := new(Expiration).SetString(str)
	if decoded == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	if exp.String() != decoded.String() {
		t.Errorf("expected %v, got %v", exp, decoded)
	}
	if exp := new(Expiration).SetString("invalid"); exp != nil {
		t.Errorf("expected nil, got %v", exp)
	}
	if exp := new(Expiration).String(); exp != "" {
		t.Errorf("expected empty string, got %v", exp)
	}
}

func TestBytesSetBytesExpiration(t *testing.T) {
	exp := NewExpiration(minDuration)
	b := exp.Bytes()
	decoded := new(Expiration).SetBytes(b)
	if decoded == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	if !bytes.Equal(b, decoded.Bytes()) {
		t.Errorf("expected %v, got %v", exp, decoded)
	}
	if exp := new(Expiration).SetBytes([]byte("invalid")); exp != nil {
		t.Errorf("expected nil, got %v", exp)
	}
	if exp := new(Expiration).Bytes(); exp != nil {
		t.Errorf("expected nil, got %v", exp)
	}
}

func TestMarshalUnmarshalExpiration(t *testing.T) {
	exp := NewExpiration(minDuration)
	encoded := exp.Marshal()
	decoded := new(Expiration).Unmarshal(encoded)
	if decoded == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	if exp.String() != decoded.String() {
		t.Errorf("expected %v, got %v", exp, decoded)
	}
	if res := new(Expiration).Marshal(); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
	if res := new(Expiration).Unmarshal([]byte{1}); res != nil {
		t.Errorf("expected nil, got %v", res)
	}
}
