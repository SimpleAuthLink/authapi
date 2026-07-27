package token

import (
	"bytes"
	"testing"
	"time"
)

func TestStringSetStringToken(t *testing.T) {
	var token *Token
	token.SetString("test")
	if token.String() != "" {
		t.Errorf("expected empty string, got %s", token.String())
	}
	token = token.SetString("test")
	if token.String() != "" {
		t.Errorf("expected empty string, got %s", token.String())
	}
	exp := new(Expiration).SetDuration(minDuration * 2)
	expected := string(exp.Marshal()) + string(tokenSeparator)
	token.SetString(expected)
	if token.String() != expected {
		t.Errorf("expected %s, got %s", expected, token.String())
	}

	expected = string(tokenSeparator) + "testSignature"
	token.SetString(expected)
	if token.String() != expected {
		t.Errorf("expected %s, got %s", expected, token.String())
	}

	expected = string(exp.Marshal()) + string(tokenSeparator) + "testSignature"
	token.SetString(expected)
	if token.String() != expected {
		t.Errorf("expected %s, got %s", expected, token.String())
	}
}

func TestBytesSetBytesToken(t *testing.T) {
	var token *Token
	token.SetBytes([]byte("test"))
	if token.Bytes() != nil {
		t.Errorf("expected nil, got %v", token.Bytes())
	}
	token = token.SetBytes([]byte("test"))
	if token.Bytes() != nil {
		t.Errorf("expected nil, got %v", token.Bytes())
	}
	exp := new(Expiration).SetDuration(minDuration * 2)
	onlyExp := append(exp.Marshal(), tokenSeparator)
	token.SetBytes(onlyExp)
	if !bytes.Equal(token.Bytes(), onlyExp) {
		t.Errorf("expected %v, got %v", onlyExp, token.Bytes())
	}

	onlySign := append([]byte{tokenSeparator}, []byte("testSignature")...)
	token.SetBytes(onlySign)
	if !bytes.Equal(token.Bytes(), onlySign) {
		t.Errorf("expected %v, got %v", onlySign, token.Bytes())
	}

	fullToken := append(append(exp.Marshal(), tokenSeparator), []byte("testSignature")...)
	token.SetBytes(fullToken)
	if !bytes.Equal(token.Bytes(), fullToken) {
		t.Errorf("expected %v, got %v", fullToken, token.Bytes())
	}
}

func TestExpirationSetExpirationToken(t *testing.T) {
	exp := new(Expiration).SetDuration(minDuration * 2)
	var token *Token
	token.SetExpiration(*exp)
	if token.String() != "" {
		t.Errorf("expected empty string, got %s", token.String())
	}
	token = token.SetExpiration(Expiration(time.Now().Add(-time.Second)))
	if token.String() != "" {
		t.Errorf("expected empty string, got %s", token.String())
	}
	token.SetExpiration(*exp)
	justExpExp := token.Expiration()
	if justExpExp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	if justExpExp.String() != exp.String() {
		t.Errorf("expected %v, got %v", exp, justExpExp)
	}
	token.SetSignature([]byte("test"))
	completeTokenExp := token.Expiration()
	if completeTokenExp == nil {
		t.Fatalf("expected valid expiration, got nil")
	}
	if completeTokenExp.String() != exp.String() {
		t.Errorf("expected %v, got %v", exp, completeTokenExp)
	}
	exp = new(Expiration).SetDuration(minDuration * 3)
	token.SetExpiration(*exp)
	newExpExp := token.Expiration()
	if newExpExp == nil || newExpExp.String() != exp.String() {
		t.Errorf("expected %v, got %v", exp, newExpExp)
	}
	validStr := token.String()
	newtoken := new(Token).SetString(validStr)
	newtokenExp := newtoken.Expiration()
	if newtokenExp == nil || newtokenExp.String() != exp.String() {
		t.Errorf("expected %v, got %v", exp, newtokenExp)
	}
}

func TestSignatureSetSignatureToken(t *testing.T) {
	var token *Token
	token.SetSignature([]byte("test"))
	if token.String() != "" {
		t.Errorf("expected empty string, got %s", token.String())
	}
	token = token.SetSignature([]byte("test"))
	justSignSign := token.Signature()
	if !bytes.Equal(justSignSign, []byte("test")) {
		t.Errorf("expected %v, got %v", []byte("test"), justSignSign)
	}
	token.SetExpiration(*new(Expiration).SetDuration(minDuration * 3))
	completeTokenSign := token.Signature()
	if !bytes.Equal(completeTokenSign, []byte("test")) {
		t.Errorf("expected %v, got %v", []byte("test"), completeTokenSign)
	}
	validStr := token.String()
	newtoken := new(Token).SetString(validStr)
	newtokenSign := newtoken.Signature()
	if !bytes.Equal(newtokenSign, []byte("test")) {
		t.Errorf("expected %v, got %v", []byte("test"), newtokenSign)
	}
}

func Test_partsToken(t *testing.T) {
	var token *Token
	if _, _, ok := token.parts(); ok {
		t.Errorf("expected false, got true")
	}
	exp := new(Expiration).SetDuration(minDuration * 2)
	token = token.SetExpiration(*exp)
	rawExp, _, ok := token.parts()
	if !ok {
		t.Errorf("expected true, got false")
	}
	if !bytes.Equal(rawExp, exp.Marshal()) {
		t.Errorf("expected %v, got %v", exp.Marshal(), rawExp)
	}
	token.SetSignature([]byte("test"))
	rawExp, sign, ok := token.parts()
	if !ok {
		t.Errorf("expected true, got false")
	}
	if !bytes.Equal(rawExp, exp.Marshal()) {
		t.Errorf("expected %v, got %v", exp.Marshal(), rawExp)
	}
	if !bytes.Equal(sign, []byte("test")) {
		t.Errorf("expected %v, got %v", []byte("test"), sign)
	}
}
