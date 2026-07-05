package client

import "go.k7z7z.cc/x/errors"

var (
	ErrInvalidAPIEndpoint        = errors.New("invalid API endpoint")
	ErrInvalidTimeout            = errors.New("invalid timeout value")
	ErrInvalidAppID              = errors.New("invalid application ID")
	ErrInvalidAppName            = errors.New("invalid application ID name")
	ErrInvalidAppRedirectURI     = errors.New("invalid application redirect URI")
	ErrInvalidAppSessionDuration = errors.New("invalid application session duration")
	ErrInvalidAppSecret          = errors.New("invalid application secret")
	ErrInvalidEmailAddress       = errors.New("invalid email address")
	ErrInvalidRequestTokenInputs = errors.New("invalid request token inputs")
	ErrAPIUnavailable            = errors.New("API is unavailable")
	ErrRequestAppID              = errors.New("failed to request application ID")
	ErrRequestToken              = errors.New("failed to request token")
	ErrInvalidToken              = errors.New("invalid token provided")
	ErrCreateRequest             = errors.New("failed to create request")
	ErrMissingAuthHeaders        = errors.New("missing authentication headers")
	ErrMissingAuthURLParams      = errors.New("missing authentication URL parameters")
)
