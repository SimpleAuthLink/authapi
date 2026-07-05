package api

import (
	"net/mail"
	"time"

	"github.com/simpleauthlink/authapi/token"
)

type AppIDRequest struct {
	Name        string `json:"name"`
	Duration    string `json:"session_duration"`
	RedirectURL string `json:"redirect_url"`
	Secret      string `json:"secret"`
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
	Email string `json:"email"`
}

func (tr *TokenRequest) IsEmail() bool {
	_, err := mail.ParseAddress(tr.Email)
	return err == nil
}

type TokenStatusRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type TokenStatusResponse struct {
	Valid      bool      `json:"valid"`
	Expiration time.Time `json:"expiration"`
}
