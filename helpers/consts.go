package helpers

const (
	// TokenSeparator constant is the separator used to split the token into
	// parts. It is a string with a value of "-".
	TokenSeparator = "-"
	// TokenQueryParam constant is the query parameter used to send the token in
	// the request. It is a string with a value of "token".
	TokenQueryParam = "token"
	// AppSecretHeader constant is the header used to send the app secret in the
	// request. It is a string with a value of "APP_SECRET".
	AppSecretHeader = "APP_SECRET"
	// DefaultAPIEndpoint constant is the default API endpoint used by the
	// client. It is a string with a value of "https://api.simpleauth.link/".
	DefaultAPIEndpoint = "https://api.simpleauth.link/"
	// HealthCheckPath constant is the path used to check the health of the API
	// server. It is a string with a value of "/health".
	HealthCheckPath = "/health"
	// AppEndpointPath constant is the path used to API endpoints related to
	// apps. It is a string with a value of "/app".
	AppEndpointPath = "/app"
	// UserEndpointPath constant is the path used to API endpoints related to
	// users. It is a string with a value of "/user".
	UserEndpointPath = "/user"
	// MinTokenDuration constant is the minimum duration allowed for a token to
	// be valid, which is an integer with a value of 60 (seconds).
	MinTokenDuration = 60 // seconds
	// defaultUsersQuota constant is the default number of users allowed for an
	// app, which is an integer with a value of 100.
	DefaultUsersQuota = 100 // users
	// UserIdSize constant is the size of the user id, which is an integer with a
	// value of 4 (bytes).
	UserIdSize = 4
	// AppIdSize constant is the size of the app id, which is an integer with a
	// value of 8 (bytes).
	AppIdSize = 8
	// EmailHashSize constant is the size of the email hash, which is an integer
	// with a value of 4 (bytes). The email hash is used to generate the user id
	// and the app id.
	EmailHashSize = 4
	// AppNonceSize constant is the size of the app nonce, which is an integer
	// with a value of 4 (bytes). The app nonce is used to generate the app id.
	AppNonceSize = 4
	// SecretSize constant is the size of the secret, which is an integer with a
	// value of 16 (bytes).
	SecretSize = 16
	// TokenSize constant is the size of the token, which is an integer with a
	// value of 8 (bytes).
	TokenSize = 8
)
