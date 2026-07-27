package token

import "go.k7z7z.cc/x/errors"

var (
	ErrInvalidApp             = errors.New("invalid app")
	ErrInvalidAppName         = errors.New("invalid app name")
	ErrInvalidRedirectURI     = errors.New("invalid redirect URI")
	ErrInvalidSessionDuration = errors.New("invalid session duration")
	ErrInvalidSecret          = errors.New("invalid secret")
)
