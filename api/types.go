package api

const (
	userTokenSubject = "Here is your magic link for '%s' 🔐"
	appTokenSubject  = "Your app '%s' is ready! 🎉"
)

// TokenRequest struct includes the required information by the API service to
// create a token, which is the email of the user. The app secret is also
// required but it is provided in the request headers.
type TokenRequest struct {
	Email       string `json:"email"`
	RedirectURL string `json:"redirect_url"`
	Duration    uint64 `json:"session_duration"`
}

// AppData struct includes the required information by the API service to
// create an app, which are the name, the email of the admin, the session
// duration and the callback URL.
type AppData struct {
	Name         string `json:"name"`
	Email        string `json:"admin_email"`
	Duration     uint64 `json:"session_duration"`
	RedirectURL  string `json:"redirect_url"`
	UsersQuota   int64  `json:"users_quota"`
	CurrentUsers int64  `json:"current_users"`
}
