package token

import "fmt"

var (
	ErrInvalidAppID           = fmt.Errorf("invalid app ID")
	ErrInvalidAppName         = fmt.Errorf("invalid app name")
	ErrInvalidRedirectURI     = fmt.Errorf("invalid redirect URI")
	ErrInvalidSessionDuration = fmt.Errorf("invalid session duration")
)
