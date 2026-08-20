package token

import (
	"bytes"
	"testing"

	"go.k7z7z.cc/x/encoding/base64url"
)

func TestEmailBytes(t *testing.T) {
	testEmail := []byte("user@example.com")
	t.Run("existing email pointer", func(t *testing.T) {
		email := new(Email).SetBytes(testEmail)
		if !bytes.Equal(email.Bytes(), testEmail) {
			t.Fatalf("expected %s, got %s", string(testEmail), string(email.Bytes()))
		}
	})

	t.Run("nil email pointer", func(t *testing.T) {
		var email *Email
		email = email.SetBytes(testEmail)
		if !bytes.Equal(email.Bytes(), testEmail) {
			t.Fatalf("expected %s, got %s", string(testEmail), string(email.Bytes()))
		}
	})
}

func TestEmailMarshal(t *testing.T) {
	t.Run("success marshal unmarshal", func(t *testing.T) {
		testEmail := []byte("user@example.com")
		expectedEmail := new(Email).SetBytes(testEmail)

		enc := expectedEmail.Marshal()
		newEmail := new(Email).Unmarshal(enc)
		if !bytes.Equal(newEmail.Bytes(), expectedEmail.Bytes()) {
			t.Fatalf("expected %s, got %s", string(expectedEmail.Bytes()), string(newEmail.Bytes()))
		}
	})

	t.Run("marshal nil email", func(t *testing.T) {
		var nilEmail *Email
		enc := nilEmail.Marshal()
		if enc != nil {
			t.Fatalf("expected nil, got %v", enc)
		}
	})

	t.Run("marshal empty email", func(t *testing.T) {
		enc := new(Email).Marshal()
		if enc != nil {
			t.Fatalf("expected nil, got %v", enc)
		}
	})

	t.Run("unmarshal nil email", func(t *testing.T) {
		rawEmail := "user@example.com"
		encEmail := base64url.RawEncode([]byte(rawEmail))
		var nilEmail *Email
		newEmail := nilEmail.Unmarshal(encEmail)
		if !bytes.Equal(newEmail.Bytes(), []byte(rawEmail)) {
			t.Fatalf("expected %s, got %s", rawEmail, string(newEmail.Bytes()))
		}
	})

	t.Run("unmarshal wrong email", func(t *testing.T) {
		enc := new(Email).Unmarshal([]byte("SGVsbG8==="))
		if len(enc.Bytes()) != 0 {
			t.Fatalf("expected nil, got %v", enc.Bytes())
		}
	})
}

func TestEmailStrin(t *testing.T) {
	rawEmail := "user@example.com"
	newEmail := new(Email).SetString(rawEmail)
	if newEmail.String() != rawEmail {
		t.Fatalf("expected %s, got %s", rawEmail, newEmail.String())
	}
}
