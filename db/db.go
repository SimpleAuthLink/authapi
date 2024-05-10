package db

import (
	"fmt"
	"time"
)

var (
	ErrInvalidConfig = fmt.Errorf("invalid database config")
	ErrOpenConn      = fmt.Errorf("error opening database")
	ErrCloseConn     = fmt.Errorf("error closing database")
	ErrAppNotFound   = fmt.Errorf("app not found")
	ErrGetApp        = fmt.Errorf("error getting the app from database")
	ErrSetApp        = fmt.Errorf("error storing the app in database")
	ErrDelApp        = fmt.Errorf("error deleting the app from database")
	ErrSetSecret     = fmt.Errorf("error storing the secret in database")
	ErrDelSecret     = fmt.Errorf("error deleting the secret from database")
	ErrTokenNotFound = fmt.Errorf("token not found")
	ErrGetToken      = fmt.Errorf("error getting the token from database")
	ErrSetToken      = fmt.Errorf("error storing the token in database")
	ErrDelToken      = fmt.Errorf("error deleting the token from database")
)


type App struct {
	Name            string
	AdminEmail      string
	SessionDuration int64
	Callback        string
}

type Token string

type DB interface {
	Init(config any) error
	Close() error
	AppById(appId string) (*App, error)
	AppBySecret(secret string) (*App, string, error)
	SetApp(appId string, app *App) error
	DeleteApp(appId string) error
	SetSecret(secret, appId string) error
	DeleteSecret(secret string) error
	TokenExpiration(token Token) (time.Time, error)
	SetToken(token Token, expiration time.Time) error
	DeleteToken(token Token) error
}
