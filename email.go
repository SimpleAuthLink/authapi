package authapi

import (
	"fmt"
	"net/smtp"
)

type EmailConfig struct {
	Address   string
	EmailHost string
	EmailPort int
	Password  string
}

type Email struct {
	To      string
	Subject string
	Body    string
}

const plainTemplate = "Subject: %s\r\n\r\n%s\r\n"

func (e *Email) Send(conf *EmailConfig) error {
	// encode the message with the template, the subject and the body
	msg := []byte(fmt.Sprintf(plainTemplate, e.Subject, e.Body))
	// create the auth object with the email credentials
	auth := smtp.PlainAuth("", conf.Address, conf.Password, conf.EmailHost)
	// create the server string with the host and the port and the receipts
	server := fmt.Sprintf("%s:%d", conf.EmailHost, conf.EmailPort)
	receipts := []string{e.To}
	// send the email
	if err := smtp.SendMail(server, auth, conf.Address, receipts, msg); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}
