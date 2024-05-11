package authapi

import (
	"encoding/hex"
	"fmt"

	"github.com/lucasmenendez/authapi/db"
)

// minDuration constant is the minimum duration allowed for a token to be valid,
// which is an integer with a value of 60 (seconds).
const minDuration = 60 // seconds

// authApp method creates a new app based on the provided name, email, callback
// and duration. It returns the app id and the app secret. If the name, email or
// callback are empty, it returns an error. If the duration is less than the
// minimum duration, it returns an error. If something fails during the process,
// it returns an error. The app id and the app secret are generated based on the
// email using the generateApp function. The app is stored in the database using
// the app id as the key. The secret is stored in the database using the hashed
// secret as the key. The hashed secret is required to be compared with the
// secret provided by the user in the requests.
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

// generateApp function generates an app based on the email. It returns the app
// id, the app secret and the hashed secret. If the email is empty or something
// fails during the process, it returns an error. The app id is generated
// hashing the email with a length of 4 bytes. The app secret is a random
// sequence of 16 bytes encoded as a hexadecimal string. The hashed secret is
// required to store the secret in the database without exposing it.
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