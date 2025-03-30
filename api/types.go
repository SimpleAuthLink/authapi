package api

import (
	"time"

	"github.com/simpleauthlink/authapi/token"
)

type AppIDRequest struct {
	Name        string `json:"name"`
	Duration    string `json:"session_duration"`
	RedirectURL string `json:"redirect_url"`
	Secret      string `json:"secret"`
}

func (data *AppIDRequest) parseApp() *token.App {
	if duration, err := time.ParseDuration(data.Duration); err == nil {
		return &token.App{
			Name:            data.Name,
			RedirectURI:     data.RedirectURL,
			SessionDuration: duration,
		}
	}
	return new(token.App)
}

type AppIDResponse struct {
	ID string `json:"id"`
}

type TokenRequest struct {
	Email string `json:"email"`
}

type TokenStatusRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type TokenStatusResponse struct {
	Valid      bool      `json:"valid"`
	Expiration time.Time `json:"expiration"`
}
