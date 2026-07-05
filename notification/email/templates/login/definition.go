package login

import (
	_ "embed"
	"regexp"

	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/token"
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

// FindToken function extracts the token from the email content. It uses a
// regular expression to fill the template with regex and find the token in
// the email content. Then it decodes the token and returns it. If the token
// is not found, it returns nil.
func FindToken(emailAddress, content string) *token.Token {
	loginData := Data{
		AppName: `.+`,
		Email:   emailAddress,
		Token:   `(.+\..+)`,
		Link:    `.+`,
	}
	loginEmail, err := Template.Compose(email.EmailParams{
		To:      emailAddress,
		Subject: loginData.Subject(),
	}, loginData)
	if err != nil {
		return nil
	}
	tokenRgx := regexp.MustCompile(string(loginEmail.PlainBody))
	tokenResult := tokenRgx.FindAllStringSubmatch(content, -1)
	if len(tokenResult) < 1 || len(tokenResult[0]) < 2 {
		return nil
	}
	return new(token.Token).SetString(tokenResult[0][1])
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
