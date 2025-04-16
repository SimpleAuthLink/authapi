package token

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestSetPartsSecret(t *testing.T) {
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	valid := new(Secret).SetParts(servicePart, appPart)
	if valid == nil {
		t.Errorf("expected Secret, got nil")
	}
	hServicePart := sha256.Sum256(servicePart)
	hAppPart := sha256.Sum256(appPart)
	expected := append(hServicePart[:], hAppPart[:]...)
	if !bytes.Equal(valid.Bytes(), expected) {
		t.Errorf("expected %x, got %x", expected[:], valid.Bytes())
	}
	if noServicePart := new(Secret).SetParts(nil, appPart); !bytes.Equal(noServicePart.Bytes(), hAppPart[:]) {
		t.Errorf("expected nil, got %x", noServicePart)
	}
	if noAppPart := new(Secret).SetParts(servicePart, nil); !bytes.Equal(noAppPart.Bytes(), hServicePart[:]) {
		t.Errorf("expected nil, got %x", noAppPart)
	}
	var nilSecret *Secret
	valid = nilSecret.SetParts(servicePart, appPart)
	if !bytes.Equal(valid.Bytes(), expected) {
		t.Errorf("expected nil, got %v", nilSecret)
	}
}

func TestValidSecret(t *testing.T) {
	var nilSecret *Secret
	if valid := nilSecret.Valid(); valid {
		t.Errorf("expected false, got %v", valid)
	}
	if valid := new(Secret).Valid(); valid {
		t.Errorf("expected false, got %v", valid)
	}
	servicePart := []byte("service-secret")
	singlePart := new(Secret).SetParts(servicePart)
	if valid := singlePart.Valid(); valid {
		t.Errorf("expected false, got %v", valid)
	}
	appPart := []byte("app-secret")
	valid := new(Secret).SetParts(servicePart, appPart)
	if !valid.Valid() {
		t.Errorf("expected true, got false")
	}
}

func TestSecretHash(t *testing.T) {
	servicePart := []byte("service-secret")
	appPart := []byte("app-secret")
	secret := new(Secret).SetParts(servicePart, appPart)
	h := secret.Hash()
	if h == nil {
		t.Errorf("expected hash, got nil")
	}
	if len(h) != secretHashSize {
		t.Errorf("expected hash size %d, got %d", secretHashSize, len(h))
	}
	hSecret := sha256.Sum256(secret.Bytes())
	if !bytes.Equal(h, hSecret[:secretHashSize]) {
		t.Errorf("expected %x, got %x", hSecret[:secretHashSize], h)
	}
	// try to hash a nil secret
	var nilSecret *Secret
	h = nilSecret.Hash()
	if h != nil {
		t.Errorf("expected nil, got %x", h)
	}
}
