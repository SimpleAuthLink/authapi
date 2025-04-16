package api

// routes paths constants
const (
	// HealthCheckPath constant is the path used to check the health of the API
	// server. It is a string with a value of "/health".
	HealthCheckPath = "/ping"
	// AppsPath constant is the path used to create the apps in the API server.
	AppsPath = "/apps"
	// TokensPath constant is the path used to generate and verify the tokens
	// in the API server.
	TokensPath = "/tokens"
	// DemoInboxPath constant is the path used to get the demo email inbox
	// in the API server when it runs in demo mode.
	DemoInboxPath = "/demo/inbox"
)

// other api related constants
const (
	// appIDHeader constant is the header of the app ID in the request. It is
	// used to authenticate the app making the request.
	appIDHeader = "AppID"
	// appSecretHeader constant is the header of the app secret in the request
	// It is used to authenticate the app making the request.
	appSecretHeader = "AppSecret"
)
