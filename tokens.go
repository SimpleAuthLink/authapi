package authapi

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lucasmenendez/authapi/db"
)

const (
	tokenTemplate  = "%s-%s-%s"
	tokenSeparator = "-"

	userIdSize = 4
	appIdSize  = 4
	secretSize = 16
	tokenSize  = 8
)

func (s *Service) magicLink(rawSecret, email string) (string, error) {
	// check if the secret and email are not empty
	if len(rawSecret) == 0 || len(email) == 0 {
		return "", fmt.Errorf("secret and email are required")
	}
	// get app secret from raw secret
	appSecret, err := hash(rawSecret, secretSize)
	if err != nil {
		return "", err
	}
	// get app and app id from the database based on the secret
	app, appId, err := s.db.AppBySecret(appSecret)
	if err != nil {
		return "", err
	}
	// generate token and calculate expiration
	token, err := encodeUserToken(appId, email)
	if err != nil {
		return "", err
	}
	expiration := time.Now().Add(time.Duration(app.SessionDuration) * time.Second)
	// set token and expiration in the database
	if err := s.db.SetToken(db.Token(token), expiration); err != nil {
		return "", err
	}
	// return the magic link based on the app callback and the generated token
	// TODO: user net/url package
	return fmt.Sprintf("%s?token=%s", app.Callback, token), nil
}

func (s *Service) validUserToken(token string) bool {
	// check if the token is not empty
	if len(token) == 0 {
		return false
	}
	// get the app id from the token
	appId, _, err := decodeUserToken(token)
	if err != nil {
		return false
	}
	// check if the app in the database
	if _, err := s.db.AppById(appId); err != nil {
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

func encodeUserToken(appId, email string) (string, error) {
	// check if the app id and email are not empty
	if len(appId) == 0 || len(email) == 0 {
		return "", fmt.Errorf("appId and email are required")
	}
	bToken := randBytes(8)
	hexToken := hex.EncodeToString(bToken)
	// hash email
	userId, err := hash(email, 4)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(tokenTemplate, hexToken, appId, userId), nil
}

func decodeUserToken(token string) (string, string, error) {
	tokenParts := strings.Split(token, tokenSeparator)
	if len(tokenParts) != 3 {
		return "", "", fmt.Errorf("invalid token")
	}
	return tokenParts[1], tokenParts[2], nil
}
