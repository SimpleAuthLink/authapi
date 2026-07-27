package email

import (
	"bytes"
	htmltemplate "html/template"
	texttemplate "text/template"
)

// EmailTemplate is the definition of an email template, which contains the
// HTML and plain text placeholders to be filled with the data.
type EmailTemplate struct {
	HTML  string
	Plain string
}

// Compose methods fills the email template with the data and returns the email
// ready to be sent. It returns the email or an error if the template could not
// be filled. It tries to fill both the HTML and plain text templates, but if
// any of them is missing, it will return an error. If some of the placeholders
// in the template are not filled, they will be left as they are.
func (temp *EmailTemplate) Compose(params EmailParams, data any) (*EmailNotification, error) {
	if !params.Valid() {
		return nil, ErrComposeEmail
	}
	// compose the html body
	body, err := temp.composeHTML(data)
	if err != nil {
		return nil, err
	}
	// compose the plain body
	plainBody, err := temp.composePlain(data)
	if err != nil {
		return nil, err
	}
	// if both bodies are empty, return an error
	if plainBody == nil && body == nil {
		return nil, ErrInvalidTemplate
	}
	// return the email with the filled bodies
	email := &EmailNotification{
		Params:    params,
		Body:      body,
		PlainBody: plainBody,
	}
	return email, nil
}

// composePlain method fills the plain text template with the data and returns the
// filled content as a byte slice. It returns the filled template or an error
// if the template could not be filled. If the plain text template is empty, it
// returns nil and no error.
func (temp *EmailTemplate) composePlain(data any) ([]byte, error) {
	if temp.Plain == "" {
		return nil, nil
	}
	// parse the placeholder plain body template
	tmpl, err := texttemplate.New("plain").Parse(temp.Plain)
	if err != nil {
		return nil, err
	}
	// inflate the template with the data
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	// return the notification with the plain body filled with the data
	return buf.Bytes(), nil
}

// composeHTML method fills the HTML template with the data and returns the filled
// content as a byte slice. It returns the filled template or an error if the
// template could not be filled. If the HTML template is empty, it returns nil
// and no error.
func (temp *EmailTemplate) composeHTML(data any) ([]byte, error) {
	if temp.HTML == "" {
		return nil, nil
	}
	// parse the email template
	tmpl, err := htmltemplate.New("html").Parse(temp.HTML)
	if err != nil {
		return nil, err
	}
	// inflate the template with the data
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	// set the body of the notification
	return buf.Bytes(), nil
}
