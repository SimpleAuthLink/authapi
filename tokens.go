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

// magicLink function generates and returns a magic link based on the provided
// app secret and user email. If the secret or the email are empty, it returns
// an error. It gets the app id from the database based on the secret. It
// generates a token and calculates the expiration time based on the app session
// duration. It stores the token and the expiration time in the database. It
// returns the magic link composed of the app callback and the generated token.
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

// validUserToken function checks if the provided token is valid. It checks if
// the token is not empty, if the app id is in the database, if the token is not
// expired and if the token is in the database. If the token is invalid, it
// returns false. If something goes wrong during the process, it logs the error
// and returns false. If the token is valid, it returns true.
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

// encodeUserToken function encodes the user information into a token and
// returns it. It receives the app id and the email of the user and returns the
// token. If the app id or the email are empty, it returns an error. The token
// is composed of three parts separated by a token separator. The first part is
// a random sequence of 8 bytes encoded as a hexadecimal string. The second part
// is the app id and the third part is the user id. The user id is generated
// hashing the email with a length of 4 bytes. The token is returned following
// the token format:
//
//	[randomPart(16)]-[appId(8)]-[userId(8)]
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

// decodeUserToken function decodes the user information from the token provided
// and returns the app id and the user id. If the token is invalid, it returns
// an error. It splits the provided token by the token separator and returns the
// second and third parts, which are the app id and the user id respectively,
// following the token format:
//
//	[randomPart(16)]-[appId(8)]-[userId(8)]
func decodeUserToken(token string) (string, string, error) {
	tokenParts := strings.Split(token, tokenSeparator)
	if len(tokenParts) != 3 {
		return "", "", fmt.Errorf("invalid token")
	}
	return tokenParts[1], tokenParts[2], nil
}
