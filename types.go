package authapi

const (
	userTokenSubject = "Your login link is ready!"
	userTokenBody    = "Login clicking here: %s"
	appTokenSubject  = "Your app is ready!"
	appTokenBody     = "Here is the secret for your app '%s' (%s):\n\n\t%s\n\nKeep it safe!"
)

// TokenRequest struct includes the required information by the API service to
// create a token, which is the email of the user. The app secret is also
// required but it is provided in the request headers.
type TokenRequest struct {
	Email string `json:"email"`
}

// AppRequest struct includes the required information by the API service to
// create an app, which are the name, the email of the admin, the session
// duration and the callback URL.
type AppRequest struct {
	Name     string `json:"name"`
	Email    string `json:"admin_email"`
	Duration int64  `json:"session_duration"`
	Callback string `json:"callback"`
}
