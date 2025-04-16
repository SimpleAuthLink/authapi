package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppConfigFromRequest(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedAppID  string
		expectedSecret string
		expectError    bool
	}{
		{
			name:           "Valid headers",
			headers:        map[string]string{appIDHeader: "testAppID", appSecretHeader: "testAppSecret"},
			expectedAppID:  "testAppID",
			expectedSecret: "testAppSecret",
			expectError:    false,
		},
		{
			name:        "Missing app id",
			headers:     map[string]string{appSecretHeader: "testAppSecret"},
			expectError: true,
		},
		{
			name:        "Missing app secret",
			headers:     map[string]string{appIDHeader: "testAppID"},
			expectError: true,
		},
		{
			name:        "Missing both headers",
			headers:     map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			appID, appSecret, err := appConfigFromRequest(req)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if appID != tt.expectedAppID {
				t.Errorf("expected appID: %s, got: %s", tt.expectedAppID, appID)
			}
			if appSecret != tt.expectedSecret {
				t.Errorf("expected appSecret: %s, got: %s", tt.expectedSecret, appSecret)
			}
		})
	}
}
