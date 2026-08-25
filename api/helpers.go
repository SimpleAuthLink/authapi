package api

import (
	"fmt"
	"net/http"

	"github.com/simpleauthlink/authapi/token"
	"go.k7z7z.cc/x/net/http/io"
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

func appAndSecretFromRequest(r *http.Request, serviceSecret []byte) (*token.App, *token.Secret, *io.APIError) {
	// get the app id from the request header
	strAppID, strAppSecret, err := appConfigFromRequest(r)
	if err != nil {
		return nil, nil, ErrInvalidAppHeaders
	}
	// decode the app id get the app from it
	appID := new(token.AppID).SetString(strAppID)
	app := new(token.App).SetID(appID)
	// compose the app secret with both parts
	secret := new(token.Secret).SetParts(serviceSecret, []byte(strAppSecret))
	if !secret.Valid() {
		return nil, nil, ErrInvalidAppSecret
	}
	// check if the app id is valid (it should be a valid app)
	if err := app.Valid(secret.Hash()); err != nil {
		return nil, nil, ErrInvalidAppID
	}
	return app, secret, nil
}
