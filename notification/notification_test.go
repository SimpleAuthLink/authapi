package notification

import (
	"testing"
)

func TestNotificationParams_Valid(t *testing.T) {
	tests := []struct {
		name   string
		params NotificationParams
		valid  bool
	}{
		{
			name:   "Valid params",
			params: NotificationParams{To: "test@example.com", Subject: "Test Subject"},
			valid:  true,
		},
		{
			name:   "Invalid email",
			params: NotificationParams{To: "invalid-email", Subject: "Test Subject"},
			valid:  false,
		},
		{
			name:   "Empty subject",
			params: NotificationParams{To: "test@example.com", Subject: ""},
			valid:  false,
		},
		{
			name:   "Empty email and subject",
			params: NotificationParams{To: "", Subject: ""},
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
		notification Notification
		valid        bool
	}{
		{
			name: "Valid notification with Body",
			notification: Notification{
				Params: NotificationParams{To: "test@example.com", Subject: "Test Subject"},
				Body:   []byte("Test Body"),
			},
			valid: true,
		},
		{
			name: "Valid notification with PlainBody",
			notification: Notification{
				Params:    NotificationParams{To: "test@example.com", Subject: "Test Subject"},
				PlainBody: []byte("Test Plain Body"),
			},
			valid: true,
		},
		{
			name: "Invalid notification with empty Body and PlainBody",
			notification: Notification{
				Params: NotificationParams{To: "test@example.com", Subject: "Test Subject"},
			},
			valid: false,
		},
		{
			name: "Invalid notification with invalid Params",
			notification: Notification{
				Params: NotificationParams{To: "invalid-email", Subject: "Test Subject"},
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
