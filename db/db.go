package db

import (
	"fmt"
	"time"
)

var (
	// ErrInvalidConfig error is returned when the provided database
	// configuration is missing or invalid.
	ErrInvalidConfig = fmt.Errorf("invalid database config")
	// ErrOpenConn error is returned when the database connection can't be
	// opened with the provided configuration.
	ErrOpenConn = fmt.Errorf("error opening database")
	// ErrCloseConn error is returned when the database connection can't be
	// closed.
	ErrCloseConn = fmt.Errorf("error closing database")
	// ErrAppNotFound error is returned when the desired app is not found in the
	// database.
	ErrAppNotFound = fmt.Errorf("app not found")
	// ErrGetApp error is returned when something fails getting a app from the
	// database.
	ErrGetApp = fmt.Errorf("error getting the app from database")
	// ErrSetApp error is returned when something fails storing a app in the
	// database.
	ErrSetApp = fmt.Errorf("error storing the app in database")
	// ErrDelApp error is returned when something fails deleting a app from the
	// database.
	ErrDelApp = fmt.Errorf("error deleting the app from database")
	// ErrSecretNotFound error is returned when the desired secret is not found
	// in the database.
	ErrSetSecret = fmt.Errorf("error storing the secret in database")
	// ErrDelSecret error is returned when something fails deleting a secret
	// from the database.
	ErrDelSecret = fmt.Errorf("error deleting the secret from database")
	// ErrTokenNotFound error is returned when the desired token is not found in
	// the database.
	ErrTokenNotFound = fmt.Errorf("token not found")
	// ErrGetToken error is returned when something fails getting a token from
	// the database.
	ErrGetToken = fmt.Errorf("error getting the token from database")
	// ErrSetToken error is returned when something fails storing a token in the
	// database.
	ErrSetToken = fmt.Errorf("error storing the token in database")
	// ErrDelToken error is returned when something fails deleting a token from
	// the database.
	ErrDelToken = fmt.Errorf("error deleting the token from database")
)

// App struct represents the application information that is stored in the
// database.
type App struct {
	Name            string
	AdminEmail      string
	SessionDuration int64
	RedirectURL     string
}

// Token type represents the token that is stored in the database.
type Token string

type DB interface {
	// Init method allows to the interface implementation to receive some config
	// information and init the database connection. It returns an error if the
	// config is invalid or the connection can't be opened.
	Init(config any) error
	// Close method allows to the interface implementation to close the database
	// connection. It returns an error if something fails during the closing.
	Close() error
	// AppById method gets an app from the database based on the app id. It
	// returns the app and an error if something goes wrong.
	AppById(appId string) (*App, error)
	// AppBySecret method gets an app from the database based on the app secret.
	// It returns the app, the app id and an error if something goes wrong.
	AppBySecret(secret string) (*App, string, error)
	// SetApp method stores an app in the database. It returns an error if
	// something goes wrong.
	SetApp(appId string, app *App) error
	// DeleteApp method deletes an app from the database. It returns an error if
	// something goes wrong.
	DeleteApp(appId string) error
	// ValidSecret method checks if a secret is valid. It returns true if the
	// secret is valid and false if it is not.
	ValidSecret(secret, appId string) (bool, error)
	// SetSecret method stores a secret in the database. It returns an error if
	// something goes wrong.
	SetSecret(secret, appId string) error
	// DeleteSecret method deletes a secret from the database. It returns an
	// error if something goes wrong.
	DeleteSecret(secret string) error
	// TokenExpiration method gets the token expiration from the database. It
	// returns the expiration time and an error if something goes wrong.
	TokenExpiration(token Token) (time.Time, error)
	// SetToken method stores a token in the database with an expiration time.
	// It returns an error if something goes wrong.
	SetToken(token Token, expiration time.Time) error
	// DeleteToken method deletes a token from the database. It returns an error
	// if something goes wrong.
	DeleteToken(token Token) error
	// DeleteTokenByPrefix method deletes all the tokens with the provided
	// prefix from the database. It returns an error if something goes wrong.
	DeleteTokensByPrefix(prefix string) error
	// DeleteExpiredTokens method deletes all the expired tokens from the
	// database. It returns an error if something goes wrong.
	DeleteExpiredTokens() error
}
