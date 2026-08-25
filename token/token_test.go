package token

import (
	"bytes"
	"testing"
	"time"
)

func TestValid(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var nilToken *Token
		if nilToken.Valid() {
			t.Fatalf("expected invalid token, got valid one")
		}
	})

	t.Run("invalid parts", func(t *testing.T) {
		var emptyToken Token
		if emptyToken.Valid() {
			t.Fatalf("expected invalid token, got valid one")
		}
	})

	t.Run("no expiration", func(t *testing.T) {
		token := new(Token)
		token.SetSignature([]byte("abcdef1234567890"))
		token.SetEmail(*new(Email).SetString("test@email.com"))
		if token.Valid() {
			t.Fatalf("expected invalid token, got valid one")
		}
	})

	t.Run("no signature", func(t *testing.T) {
		token := new(Token)
		token.SetExpiration(*new(Expiration).SetDuration(minDuration))
		token.SetEmail(*new(Email).SetString("test@email.com"))
		if token.Valid() {
			t.Fatalf("expected invalid token, got valid one")
		}
	})

	t.Run("no email", func(t *testing.T) {
		token := new(Token)
		token.SetSignature([]byte("abcdef1234567890"))
		token.SetExpiration(*new(Expiration).SetDuration(minDuration))
		if token.Valid() {
			t.Fatalf("expected invalid token, got valid one")
		}
	})

	t.Run("valid token", func(t *testing.T) {
		token := new(Token)
		token.SetSignature([]byte("abcdef1234567890"))
		token.SetExpiration(*new(Expiration).SetDuration(minDuration))
		token.SetEmail(*new(Email).SetString("test@email.com"))
		if !token.Valid() {
			t.Fatalf("expected valid token, got invalid one")
		}
	})
}

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
	expected := string(exp.Marshal()) + string(tokenSeparator) + string(tokenSeparator)
	token.SetString(expected)
	if token.String() != expected {
		t.Errorf("expected %s, got %s", expected, token.String())
	}

	expected = string(tokenSeparator) + "testSignature" + string(tokenSeparator)
	token.SetString(expected)
	if token.String() != expected {
		t.Errorf("expected %s, got %s", expected, token.String())
	}

	expected = string(exp.Marshal()) + string(tokenSeparator) + "testSignature" + string(tokenSeparator)
	token.SetString(expected)
	if token.String() != expected {
		t.Errorf("expected %s, got %s", expected, token.String())
	}

	expected = string(exp.Marshal()) + string(tokenSeparator) + "testSignature" + string(tokenSeparator) + "testEmail"
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
	onlyExp = append(onlyExp, tokenSeparator)
	token.SetBytes(onlyExp)
	if !bytes.Equal(token.Bytes(), onlyExp) {
		t.Errorf("expected %v, got %v", onlyExp, token.Bytes())
	}

	onlySign := append([]byte{tokenSeparator}, []byte("testSignature")...)
	onlySign = append(onlySign, tokenSeparator)
	token.SetBytes(onlySign)
	if !bytes.Equal(token.Bytes(), onlySign) {
		t.Errorf("expected %v, got %v", onlySign, token.Bytes())
	}

	onlyExpAndSign := append(append(exp.Marshal(), tokenSeparator), []byte("testSignature")...)
	onlyExpAndSign = append(onlyExpAndSign, tokenSeparator)
	token.SetBytes(onlyExpAndSign)
	if !bytes.Equal(token.Bytes(), onlyExpAndSign) {
		t.Errorf("expected %v, got %v", onlyExpAndSign, token.Bytes())
	}

	fullToken := append(append(exp.Marshal(), tokenSeparator), []byte("testSignature")...)
	fullToken = append(append(fullToken, tokenSeparator), []byte("testEmail")...)
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
	if _, _, _, ok := token.split(); ok {
		t.Errorf("expected false, got true")
	}
	exp := new(Expiration).SetDuration(minDuration * 2)
	token = token.SetExpiration(*exp)
	rawExp, _, _, ok := token.split()
	if !ok {
		t.Errorf("expected true, got false")
	}
	if !bytes.Equal(rawExp, exp.Marshal()) {
		t.Errorf("expected %v, got %v", exp.Marshal(), rawExp)
	}
	token.SetSignature([]byte("test"))
	rawExp, sign, _, ok := token.split()
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

func TestSplitJoin(t *testing.T) {
	t.Run("nil token", func(t *testing.T) {
		var nilToken *Token
		if _, _, _, ok := nilToken.split(); ok {
			t.Fatalf("expected wrong split token, got ok")
		}

		if newToken := nilToken.join([]byte{}); len(newToken.Bytes()) != 0 {
			t.Fatalf("expected nil token, got %x", newToken)
		}
	})

	t.Run("invalid parts of token", func(t *testing.T) {
		var token Token = []byte("test.token")
		if _, _, _, ok := token.split(); ok {
			t.Fatalf("expected wrong split token, got ok")
		}

		if newToken := token.join([]byte{}); len(newToken.Bytes()) != 0 {
			t.Fatalf("expected nil token, got %x", newToken)
		}
	})

	t.Run("valid token", func(t *testing.T) {
		expSig := []byte("abcdef1234567890")
		expEmail := new(Email).SetString("test@email.com")
		expExpiration := new(Expiration).SetDuration(minDuration)

		token := new(Token)
		token.SetSignature(expSig)
		token.SetExpiration(*expExpiration)
		token.SetEmail(*expEmail)

		exp, sig, email, ok := token.split()
		if !ok {
			t.Fatal("expected valid token parts")
		}
		if !bytes.Equal(expSig, sig) {
			t.Fatalf("expected signature %x, got %x", expSig, sig)
		}
		if !bytes.Equal(expExpiration.Marshal(), exp) {
			t.Fatalf("expected expiration %x, got %x", expExpiration.Marshal(), exp)
		}
		if !bytes.Equal(expEmail.Marshal(), email) {
			t.Fatalf("expected email %x, got %x", expEmail.Marshal(), email)
		}

		newToken := new(Token).join(exp, sig, email)
		if !bytes.Equal(token.Bytes(), newToken.Bytes()) {
			t.Fatalf("expected token %x, got %x", token, newToken)
		}
	})
}
