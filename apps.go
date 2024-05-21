package authapi

import (
	"encoding/hex"
	"fmt"

	"github.com/simpleauthlink/authapi/db"
)

const (
	// minDuration constant is the minimum duration allowed for a token to be
	// valid, which is an integer with a value of 60 (seconds).
	minDuration = 60 // seconds
	// defaultUsersQuota constant is the default number of users allowed for an
	// app, which is an integer with a value of 25.
	defaultUsersQuota = 25 // users
)

// authApp method creates a new app based on the provided name, email, redirectURL
// and duration. It returns the app id and the app secret. If the name, email or
// redirectURL are empty, it returns an error. If the duration is less than the
// minimum duration, it returns an error. If something fails during the process,
// it returns an error. The app id and the app secret are generated based on the
// email using the generateApp function. The app is stored in the database using
// the app id as the key. The secret is stored in the database using the hashed
// secret as the key. The hashed secret is required to be compared with the
// secret provided by the user in the requests.
func (s *Service) authApp(name, email, redirectURL string, duration int64) (string, string, error) {
	// check if the name, email, and redirectURL are not empty
	if len(name) == 0 || len(email) == 0 || len(redirectURL) == 0 {
		return "", "", fmt.Errorf("name, email, and redirectURL are required")
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
		RedirectURL:     redirectURL,
		UsersQuota:      defaultUsersQuota,
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

// appMetadata method retrieves the app data based on the app id. If the app id is
// empty, it returns an error. If something fails during the process, it returns
// an error. The app data includes the name, the email of the admin, the redirect
// URL, the duration, the users quota, and the current users. The current users
// are retrieved from the database using the app id to count the number of tokens
// for the app.
func (s *Service) appMetadata(appId string) (AppData, error) {
	dbApp, err := s.db.AppById(appId)
	if err != nil {
		return AppData{}, err
	}
	app := AppData{
		Name:        dbApp.Name,
		Email:       dbApp.AdminEmail,
		RedirectURL: dbApp.RedirectURL,
		Duration:    dbApp.SessionDuration,
		UsersQuota:  dbApp.UsersQuota,
	}
	// get the number of current tokens for the app, if it fails, it returns 0
	app.CurrentUsers, _ = s.db.CountTokens(appId)
	return app, nil
}

// updateAppMetadata method updates the app metadata based on the app id, name,
// redirectURL, and duration. If the app id is empty, it returns an error. If
// the duration is non zero an less than the minimum duration, it returns an
// error. If something fails during the process, it returns an error.
func (s *Service) updateAppMetadata(appId, name, redirectURL string, duration int64) error {
	// check if the app id is not empty
	if len(appId) == 0 {
		return fmt.Errorf("app id is required")
	}
	// check if the duration is valid
	if duration != 0 && duration < minDuration {
		return fmt.Errorf("duration must be at least %d seconds", minDuration)
	}
	// get app from the database
	app, err := s.db.AppById(appId)
	if err != nil {
		return err
	}
	// update app metadata
	if name != "" {
		app.Name = name
	}
	if redirectURL != "" {
		app.RedirectURL = redirectURL
	}
	if duration != 0 {
		app.SessionDuration = duration
	}
	// store app in the database
	return s.db.SetApp(appId, app)
}

// removeApp method removes an app based on the app id. If the app id is empty,
// it returns an error. If something fails during the process, it returns an
// error. It also removes all the tokens for the app from the database using
// the app id as the prefix to find them.
func (s *Service) removeApp(appId string) error {
	// check if the app id is not empty
	if len(appId) == 0 {
		return fmt.Errorf("app id is required")
	}
	// remove all the tokens for the app from the database, using the app id as
	// the prefix
	if err := s.db.DeleteTokensByPrefix(appId); err != nil {
		return err
	}
	// remove app from the database
	return s.db.DeleteApp(appId)
}

func (s *Service) validSecret(appId, rawSecret string) bool {
	secret, err := hash(rawSecret, 16)
	if err != nil {
		return false
	}
	valid, err := s.db.ValidSecret(secret, appId)
	if err != nil {
		return false
	}
	return valid
}

// generateApp function generates an app based on the email. It returns the app
// id, the app secret and the hashed secret. If the email is empty or something
// fails during the process, it returns an error. The app id is generated
// hashing the email with a length of 4 bytes. The app secret is generated
// using the appSecret function.
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
	secret, hSecret, err := appSecret()
	if err != nil {
		return "", "", "", err
	}
	return appId, secret, hSecret, nil
}

// appSecret function generates an new app secret. It returns the secret, the
// hashed secret and an error if something fails during the process. The secret
// is a random sequence of 16 bytes encoded as a hexadecimal string. The hashed
// secret is required to store the secret in the database without exposing it.
func appSecret() (string, string, error) {
	// generate secret
	bSecret := randBytes(16)
	secret := hex.EncodeToString(bSecret)
	// hash secret
	hSecret, err := hash(secret, 16)
	if err != nil {
		return "", "", err
	}
	return secret, hSecret, nil
}
