package api

// routes paths constants
const (
	// HealthCheckPath constant is the path used to check the health of the API
	// server. It is a string with a value of "/ping".
	HealthCheckPath = "/ping"
	// AppsPath constant is the path used to create the apps in the API server.
	AppsPath = "/apps"
	// TokensPath constant is the path used to generate and verify the tokens
	// in the API server.
	TokensPath = "/tokens"
)

// other api related constants
const (
	// AppIDHeader constant is the header of the app ID in the request. It is
	// used to authenticate the app making the request.
	AppIDHeader = "AppID"
	// AppSecretHeader constant is the header of the app secret in the request
	// It is used to authenticate the app making the request.
	AppSecretHeader = "AppSecret"
)
