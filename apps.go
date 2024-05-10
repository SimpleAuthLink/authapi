package authapi

import (
	"encoding/hex"
	"fmt"

	"github.com/lucasmenendez/authapi/db"
)

const minDuration = 60 // seconds

func (s *Service) authApp(name, email, callback string, duration int64) (string, string, error) {
	// check if the name, email, and callback are not empty
	if len(name) == 0 || len(email) == 0 || len(callback) == 0 {
		return "", "", fmt.Errorf("name, email, and callback are required")
	}
	// check if the duration is valid
	if duration < minDuration {
		return "", "", fmt.Errorf("duration must be at least %d seconds", minDuration)
	}
	// compose the app struct for the database
	appData := &db.App{
		Name:            name,
		AdminEmail:      email,
		SessionDuration: duration,
		Callback:        callback,
	}
	// generate app based on email
	appId, secret, hSecret, err := generateApp(appData.AdminEmail)
	if err != nil {
		return "", "", err
	}
	// store app in the database
	if err := s.db.SetApp(appId, appData); err != nil {
		return "", "", err
	}
	// store secret in the database
	if err := s.db.SetSecret(hSecret, appId); err != nil {
		return "", "", err
	}
	return appId, secret, nil
}

func generateApp(email string) (string, string, string, error) {
	if len(email) == 0 {
		return "", "", "", fmt.Errorf("email is required")
	}
	// hash email
	appId, err := hash(email, 4)
	if err != nil {
		return "", "", "", err
	}
	// generate secret
	bSecret := randBytes(16)
	secret := hex.EncodeToString(bSecret)
	// hash secret
	hSecret, err := hash(secret, 16)
	if err != nil {
		return "", "", "", err
	}
	return appId, secret, hSecret, nil
}
