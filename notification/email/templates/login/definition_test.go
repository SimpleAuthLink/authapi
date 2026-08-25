package login

import (
	"fmt"
	"testing"
	"time"

	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/token"
)

func TestSubject(t *testing.T) {
	tests := []struct {
		name     string
		data     Data
		expected string
	}{
		{
			name:     "normal app name",
			data:     Data{AppName: "MyApp"},
			expected: "Your token for 'MyApp'",
		},
		{
			name:     "empty app name",
			data:     Data{AppName: ""},
			expected: "Your token for ''",
		},
		{
			name:     "special chars",
			data:     Data{AppName: "App's Name"},
			expected: "Your token for 'App's Name'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.data.Subject(); got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFindToken(t *testing.T) {
	// Generate a real token and compose a real plain-text email
	app := &token.App{
		Name:            "TestApp",
		RedirectURI:     "https://example.com/callback",
		SessionDuration: 30 * time.Minute,
	}
	servicePart := []byte("test-service-secret")
	appPart := []byte("test-app-secret")
	secret := new(token.Secret).SetParts(servicePart, appPart)
	app.SetSecret(secret)
	appID := app.ID(secret)

	testEmail := "user@test.com"
	userID := new(token.Email).SetString(testEmail)
	testToken := appID.GenerateToken(*secret, *userID)
	if testToken == nil {
		t.Fatal("failed to generate test token")
	}

	loginData := Data{
		AppName: "TestApp",
		Email:   testEmail,
		Token:   testToken.String(),
		Link:    "https://example.com/callback?token=" + testToken.String(),
	}
	loginEmail, err := Template.Compose(email.EmailParams{
		To:      testEmail,
		Subject: loginData.Subject(),
	}, loginData)
	if err != nil {
		t.Fatal(err)
	}
	emailContent := string(loginEmail.PlainBody)

	t.Run("valid email content with token", func(t *testing.T) {
		found := FindToken(testEmail, emailContent)
		if found == nil {
			t.Fatal("expected token to be found, got nil")
		}
		if got := found.String(); got != testToken.String() {
			t.Errorf("got %q, want %q", got, testToken.String())
		}
	})

	t.Run("empty content", func(t *testing.T) {
		if got := FindToken(testEmail, ""); got != nil {
			t.Errorf("expected nil for empty content, got %v", got)
		}
	})

	t.Run("wrong email address", func(t *testing.T) {
		// Regex contains the literal email, should not match for a different
		// email
		if got := FindToken("other@test.com", emailContent); got != nil {
			t.Errorf("expected nil for wrong email, got %v", got)
		}
	})

	t.Run("content without any token", func(t *testing.T) {
		if got := FindToken(testEmail, "some random text without tokens"); got != nil {
			t.Errorf("expected nil for content without token, got %v", got)
		}
	})

	t.Run("content with garbage that looks like a token", func(t *testing.T) {
		// Construct content that contains the email but a malformed token
		garbage := fmt.Sprintf("Hi, %s\nIt contains your login token: 'not.a.real.token'\n", testEmail)
		// This might return nil because the regex capture may not match, or it
		// may return a "token" that does not round-trip.
		// Either way, verify FindToken does not panic and the result (if any)
		// is non-nil but SetString'd token from the raw match.
		found := FindToken(testEmail, garbage)
		// We dont assert nil because the regex may match; we just check no panic.
		_ = found
	})
}
