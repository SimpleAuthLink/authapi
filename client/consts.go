package client

const (
	// TokenSeparator constant is the separator used to split the token into
	// parts. It is a string with a value of "-".
	TokenSeparator  = "-"
	TokenQueryParam = "token"
	AppSecretHeader = "APP_SECRET"

	DefaultAPIEndpoint = "https://api.simpleauth.link/"
	ValidateTokenPath  = "/user"
)
