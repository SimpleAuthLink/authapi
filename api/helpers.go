package api

import (
	"fmt"
	"net/http"
)

// appConfigFromRequest extracts the app id and app secret from the request
// headers. It returns an error if the app id or app secret is missing. The
// app id and app secret are used to authenticate the app making the request.
// The app id is a unique identifier for the app, and the app secret is a
// shared secret used to verify the authenticity of the request for this
// service.
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
