package email

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// UserEmailData struct includes the data required to fill the user email
// template.
type UserEmailData struct {
	AppName      string
	EmailHandler string
	MagicLink    string
	Token        string
}

// AppEmailData struct includes the data required to fill the app email
// template.
type AppEmailData struct {
	AppID        string
	AppName      string
	RedirectURL  string
	Secret       string
	EmailHandler string
}

// NewUserEmailData creates a new UserEmailData with the provided data.
func NewUserEmailData(appName, email, magicLink, token string) *UserEmailData {
	return &UserEmailData{
		AppName:      appName,
		EmailHandler: emailHandler(email),
		MagicLink:    magicLink,
		Token:        token,
	}
}

// NewAppEmailData creates a new AppEmailData with the provided data.
func NewAppEmailData(appID, appName, redirectURL, secret, email string) *AppEmailData {
	return &AppEmailData{
		AppID:        appID,
		AppName:      appName,
		RedirectURL:  redirectURL,
		Secret:       secret,
		EmailHandler: emailHandler(email),
	}
}

// ParseTemplate parses the template file provided with the data provided. It
// returns the parsed template as a string. If an error occurs, it returns the
// error.
func ParseTemplate(templatePath string, data interface{}) (string, error) {
	// parse the template file provided
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}
	// execute the template to fill it with the data provided
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}
	return buf.String(), nil
}

// emailHandler method extracts the email handler from the email address. It
// splits the email address by the "@" symbol and returns the first part.
func emailHandler(emailAddress string) string {
	emailParts := strings.Split(emailAddress, "@")
	return emailParts[0]
}
