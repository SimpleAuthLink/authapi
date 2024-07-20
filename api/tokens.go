package api

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/simpleauthlink/authapi/db"
	"github.com/simpleauthlink/authapi/helpers"
)

// magicLink function generates and returns a magic link, the generated token
// and the associated app name, based on the provided app secret and the user
// email. If the secret or the email are empty, it returns an error. It gets
// the app id from the database based on the secret. It generates a token and
// calculates the expiration time based on the app session duration. It stores
// the token and the expiration time in the database. It returns the magic link
// composed of the app callback and the generated token.
func (s *Service) magicLink(rawSecret, email, redirectURL string, duration uint64) (string, string, string, error) {
	// check if the secret and email are not empty
	if len(rawSecret) == 0 || len(email) == 0 {
		return "", "", "", fmt.Errorf("secret and email are required")
	}
	// get app secret from raw secret
	appSecret, err := helpers.Hash(rawSecret, helpers.SecretSize)
	if err != nil {
		return "", "", "", err
	}
	// get app and app id from the database based on the secret
	app, appId, err := s.db.AppBySecret(appSecret)
	if err != nil {
		return "", "", "", err
	}
	// get the number of tokens for the app using the app id as the prefix
	numberOfAppTokens, err := s.db.CountTokens(appId)
	if err != nil {
		return "", "", "", err
	}
	// check if the number of tokens is greater than the users quota
	if numberOfAppTokens >= app.UsersQuota {
		return "", "", "", fmt.Errorf("users quota reached")
	}
	// generate token and calculate expiration
	token, userId, err := helpers.EncodeUserToken(appId, email)
	if err != nil {
		return "", "", "", err
	}
	// by default, the session duration is the app session duration but it can
	// be overwritten by the request
	sessionDuration := app.SessionDuration
	if duration > 0 {
		sessionDuration = duration
	}
	expiration := time.Now().Add(time.Duration(sessionDuration) * time.Second)
	// check if there is a token for the user and app in the database and delete
	// it if it exists
	tokenPrefix := strings.Join([]string{appId, userId}, helpers.TokenSeparator)
	if err := s.db.DeleteTokensByPrefix(tokenPrefix); err != nil {
		if err != db.ErrTokenNotFound {
			log.Println("ERR: error checking token:", err)
		}
	}
	// set token and expiration in the database
	if err := s.db.SetToken(db.Token(token), expiration); err != nil {
		return "", "", "", err
	}
	// return the magic link based on the app callback and the generated token
	// by default, the redirect URL is the app redirect URL but it can be
	// overwritten by the request
	baseRawURL := app.RedirectURL
	if redirectURL != "" {
		baseRawURL = redirectURL
	}
	baseURL, err := url.Parse(baseRawURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid redirect URL: %w", err)
	}
	urlQuery := baseURL.Query()
	urlQuery.Set(helpers.TokenQueryParam, token)
	baseURL.RawQuery = urlQuery.Encode()

	strBaseURL := fmt.Sprintf("%s://%s", baseURL.Scheme, baseURL.Host)
	if baseURL.Path != "" {
		strBaseURL += baseURL.Path
	}
	if baseURL.Fragment != "" {
		strBaseURL += fmt.Sprintf("#%s", baseURL.Fragment)
	}
	if encoded := urlQuery.Encode(); encoded != "" {
		strBaseURL += fmt.Sprintf("?%s", encoded)
	}
	return strBaseURL, token, app.Name, nil
}

// validUserToken function checks if the provided token is valid. It checks if
// the token is not empty, if the app id is in the database, if the token is not
// expired and if the token is in the database. If the token is invalid, it
// returns false. If something goes wrong during the process, it logs the error
// and returns false. If the token is valid, it returns true.
func (s *Service) validUserToken(token, rawSecret string) bool {
	// check if the token and secret are not empty
	if len(token) == 0 || len(rawSecret) == 0 {
		return false
	}
	// get the app id from the token
	appId, _, err := helpers.DecodeUserToken(token)
	if err != nil {
		return false
	}
	// check if the secret is valid
	if !s.validSecret(appId, rawSecret) {
		return false
	}
	// get the token expiration from the database
	expiration, err := s.db.TokenExpiration(db.Token(token))
	if err != nil {
		return false
	}
	// check if the token is expired
	if time.Now().After(expiration) {
		if err := s.db.DeleteToken(db.Token(token)); err != nil {
			log.Println("ERR: error deleting token:", err)
		}
		return false
	}
	return true
}

// validAdminToken function checks if the provided token is a valid admin token.
// It checks if the token is not empty, if the app id is in the database, if the
// token is not expired and if the token is in the database. If the token is
// invalid, it returns false. It also returns the app id if the token is valid.
func (s *Service) validAdminToken(token, rawSecret string) (string, bool) {
	// check if the token and secret are not empty
	if len(token) == 0 || len(rawSecret) == 0 {
		return "", false
	}
	// get the app id from the token
	appId, userId, err := helpers.DecodeUserToken(token)
	if err != nil {
		return "", false
	}
	// the admin has the same id as the app (the hased email)
	if userId != appId {
		return "", false
	}
	// check if the secret is valid
	if !s.validSecret(appId, rawSecret) {
		return "", false
	}
	// get the token expiration from the database
	expiration, err := s.db.TokenExpiration(db.Token(token))
	if err != nil {
		return "", false
	}
	// check if the token is expired
	if time.Now().After(expiration) {
		if err := s.db.DeleteToken(db.Token(token)); err != nil {
			log.Println("ERR: error deleting token:", err)
		}
		return "", false
	}
	return appId, true
}

// sanityTokenCleaner function starts a goroutine that cleans the expired tokens
// from the database every time the cooldown time is reached. It uses a ticker
// to check the cooldown time and a context to stop the goroutine when the
// service is stopped. If something goes wrong during the process, it logs the
// error.
func (s *Service) sanityTokenCleaner() {
	s.wait.Add(1)
	go func() {
		defer s.wait.Done()
		ticker := time.NewTicker(s.cfg.CleanerCooldown)
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				if err := s.db.DeleteExpiredTokens(); err != nil {
					log.Println("ERR: error deleting expired tokens:", err)
				}
			}
		}
	}()
}
