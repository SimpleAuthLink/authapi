package email

import (
	"testing"
)

func TestEmailParams_Valid(t *testing.T) {
	tests := []struct {
		name   string
		params EmailParams
		valid  bool
	}{
		{
			name:   "Valid params",
			params: EmailParams{To: "test@example.com", Subject: "Test Subject"},
			valid:  true,
		},
		{
			name:   "Invalid email",
			params: EmailParams{To: "invalid-email", Subject: "Test Subject"},
			valid:  false,
		},
		{
			name:   "Empty subject",
			params: EmailParams{To: "test@example.com", Subject: ""},
			valid:  false,
		},
		{
			name:   "Empty email and subject",
			params: EmailParams{To: "", Subject: ""},
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.params.Valid(); got != tt.valid {
				t.Errorf("expected valid: %v, got: %v", tt.valid, got)
			}
		})
	}
}

func TestNotification_Valid(t *testing.T) {
	tests := []struct {
		name         string
		notification EmailNotification
		valid        bool
	}{
		{
			name: "Valid notification with Body",
			notification: EmailNotification{
				Params: EmailParams{To: "test@example.com", Subject: "Test Subject"},
				Body:   []byte("Test Body"),
			},
			valid: true,
		},
		{
			name: "Valid notification with PlainBody",
			notification: EmailNotification{
				Params:    EmailParams{To: "test@example.com", Subject: "Test Subject"},
				PlainBody: []byte("Test Plain Body"),
			},
			valid: true,
		},
		{
			name: "Invalid notification with empty Body and PlainBody",
			notification: EmailNotification{
				Params: EmailParams{To: "test@example.com", Subject: "Test Subject"},
			},
			valid: false,
		},
		{
			name: "Invalid notification with invalid Params",
			notification: EmailNotification{
				Params: EmailParams{To: "invalid-email", Subject: "Test Subject"},
				Body:   []byte("Test Body"),
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.notification.Valid(); got != tt.valid {
				t.Errorf("expected valid: %v, got: %v", tt.valid, got)
			}
		})
	}
}
