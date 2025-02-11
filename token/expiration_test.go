package token

import (
	"bytes"
	"testing"
	"time"
)

func TestStringSetStringExpiration(t *testing.T) {
	exp := NewExpiration(time.Second)
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
	exp := NewExpiration(time.Second)
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
	exp := NewExpiration(time.Second)
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
