package client

import "fmt"

var (
	ErrInvalidAPIEndpoint        = fmt.Errorf("invalid API endpoint")
	ErrInvalidTimeout            = fmt.Errorf("invalid timeout value")
	ErrInvalidAppID              = fmt.Errorf("invalid application ID")
	ErrInvalidAppName            = fmt.Errorf("invalid application ID name")
	ErrInvalidAppRedirectURI     = fmt.Errorf("invalid application redirect URI")
	ErrInvalidAppSessionDuration = fmt.Errorf("invalid application session duration")
	ErrInvalidAppSecret          = fmt.Errorf("invalid application secret")
	ErrInvalidEmailAddress       = fmt.Errorf("invalid email address")
	ErrAPIUnavailable            = fmt.Errorf("API is unavailable")
	ErrRequestToken              = fmt.Errorf("failed to request token")
	ErrInvalidToken              = fmt.Errorf("invalid token provided")
	ErrCreateRequest             = fmt.Errorf("failed to create request")
)
