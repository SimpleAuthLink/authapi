package api

import (
	"fmt"
	"net/http"
)

const (
	AppIDHeader     = "APP_ID"
	AppSecretHeader = "APP_SECRET"
)

func appConfigFromRequest(r *http.Request) (string, string, error) {
	// get the app id from the request header
	strAppID := r.Header.Get(AppIDHeader)
	if strAppID == "" {
		return "", "", fmt.Errorf("missing app id")
	}
	strAppSecret := r.Header.Get(AppSecretHeader)
	if strAppSecret == "" {
		return "", "", fmt.Errorf("missing app secret")
	}
	return strAppID, strAppSecret, nil
}
