package login

import (
	_ "embed"

	"github.com/simpleauthlink/authapi/email"
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

// Template is the login email template definition, which contains the HTML
// and plain text templates.
var Template = email.EmailTemplate{
	HTML: htmlTemplate,
	Plain: `Hi, {{.Email}}
You can login to '{{.AppName}}' following this link:
{{.Link}}
It contains your login token: '{{.Token}}'
Which is only valid for you and for a short period of time. 
If you didn't request this, you can ignore this email.`,
}
