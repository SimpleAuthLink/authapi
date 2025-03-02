package login

import (
	_ "embed"

	"github.com/simpleauthlink/authapi/notification/email"
)

//go:embed template.html
var htmlTemplate string

// Data struct contains the required data to fill the login email template.
type Data struct {
	AppName string
	Email   string
	Token   string
	Link    string
}

// Subject returns the email subject based on the login data.
func (d Data) Subject() string {
	return "Your token for '" + d.AppName + "'"
}

// Template is the login email template definition, which contains the HTML
// and plain text templates.
var Template = email.EmailTemplate{
	HTML: htmlTemplate,
	Plain: `Hi, {{.Email}}
You can access to '{{.AppName}}' app using the following link:
{{.Link}}
It contains your login token: '{{.Token}}'
Which is only valid for you and for a short period of time. 
If you didn't request this, you can ignore this email.`,
}
