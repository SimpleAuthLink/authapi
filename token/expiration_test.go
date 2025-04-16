package token

import (
	"bytes"
	"testing"
	"time"
)

func TestValidExpiration(t *testing.T) {
	t.Parallel()
	exp := new(Expiration).SetDuration(minDuration + time.Second)
	if !exp.Valid() {
		t.Errorf("expected valid expiration, got invalid")
	}
	time.Sleep(minDuration + (time.Second * 2))
	if exp.Valid() {
		t.Errorf("expected invalid expiration, got valid")
	}
}

func TestTimeSetTimeExpiration(t *testing.T) {
	var nilExp *Expiration
	exp := nilExp.SetTime(time.Now().Add(minDuration * 2))
	if exp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	exp = new(Expiration).SetTime(time.Now().Add(minDuration * 2))
	if exp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	expTime := exp.Time()
	expected := time.Now().Add(minDuration * 2)
	if expected.Sub(expTime) > time.Millisecond*300 {
		t.Errorf("expected %v, got %v", expected, expTime)
	}
	invalidTime := time.Now().Add(-time.Second)
	if exp := new(Expiration).SetTime(invalidTime); exp != nil {
		t.Errorf("expected nil, got %v", exp)
	}
}

func TestDurationSetDurationExpiration(t *testing.T) {
	exp := new(Expiration).SetDuration(minDuration - 1)
	if exp != nil {
		t.Fatalf("expected nil, got %v", exp)
	}
	exp = exp.SetDuration(minDuration * 2)
	if exp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	expectedDuration := time.Duration(minDuration * 2)
	if expectedDuration-exp.Duration() > time.Millisecond*300 {
		t.Errorf("expected %v, got %v", expectedDuration, exp.Duration())
	}
}

func TestStringSetStringExpiration(t *testing.T) {
	exp := new(Expiration).SetDuration(minDuration * 2)
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
	var nilExp *Expiration
	if nilExp.String() != "" {
		t.Errorf("expected empty string, got %v", nilExp.String())
	}
	if nilExp = nilExp.SetString(str); nilExp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	if nilExp.String() != str {
		t.Errorf("expected %v, got %v", str, nilExp.String())
	}
	invalidTime := time.Now().Add(-time.Second).Format(time.RFC3339Nano)
	if exp := new(Expiration).SetString(invalidTime); exp != nil {
		t.Errorf("expected nil, got %v", exp)
	}
}

func TestBytesSetBytesExpiration(t *testing.T) {
	exp := new(Expiration).SetDuration(minDuration * 2)
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
	exp := new(Expiration).SetDuration(minDuration * 2)
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
