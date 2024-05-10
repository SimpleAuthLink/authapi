package authapi

const (
	userTokenSubject = "Your login link is ready!"
	userTokenBody    = "Login clicking here: %s"
	appTokenSubject  = "Your app is ready!"
	appTokenBody     = "Here is the secret for your app '%s' (%s):\n\n\t%s\n\nKeep it safe!"
)

type TokenRequest struct {
	Email string `json:"email"`
}

type AppRequest struct {
	Name     string `json:"name"`
	Email    string `json:"admin_email"`
	Duration int64  `json:"session_duration"`
	Callback string `json:"callback"`
}
