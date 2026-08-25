package api

import (
	"time"

	"github.com/simpleauthlink/authapi/token"
)

type AppIDRequest struct {
	Name        string `json:"name" example:"MySuperMegaApp"`
	Duration    string `json:"session_duration" example:"30m"`
	RedirectURL string `json:"redirect_url" example:"https://example.com/callback"`
	Secret      string `json:"secret" example:"mysupersecret"`
}

func (data *AppIDRequest) parseApp(secret string) (*token.App, error) {
	duration, err := time.ParseDuration(data.Duration)
	if err != nil {
		return nil, err
	}
	app := &token.App{
		Name:            data.Name,
		RedirectURI:     data.RedirectURL,
		SessionDuration: duration,
	}
	finalSecret := new(token.Secret).SetParts([]byte(secret), []byte(data.Secret))
	app.SetSecret(finalSecret)
	return app, nil
}

type AppIDResponse struct {
	ID string `json:"id"`
}

type TokenRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

type TokenStatusRequest struct {
	Token string `json:"token" example:"<token>"`
}

type TokenStatusResponse struct {
	Valid      bool      `json:"valid"`
	Expiration time.Time `json:"expiration"`
}
