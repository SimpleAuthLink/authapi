package authapi

import (
	"fmt"
	"net/smtp"
)

// EmailConfig struct represents the email configuration that is needed to send
// an email using and SMTP server. It includes the email address (used as the
// sender address but also as the username for the SMTP server), the email
// server hostname, its port and the password.
type EmailConfig struct {
	Address   string
	EmailHost string
	EmailPort int
	Password  string
}

// Email struct represents the email that is going to be sent. It includes the
// recipient email address, the subject and the body of the email.
type Email struct {
	To      string
	Subject string
	Body    string
}

// plainTemplate is the template used to compose the email message. It includes
// the subject and the body of the email.
const plainTemplate = "Subject: %s\r\n\r\n%s\r\n"

// retries is the number of retries to send the email.
const retries = 3

// Send method sends the email using the provided configuration. It uses the
// email address as the sender address and the username for the SMTP server.
// It composes the email message, creates the auth object with the email
// credentials, the server string with the host and the port, and the receipts.
// Finally, it sends the email. If something fails during the process, it
// returns an error.
func (e *Email) Send(conf *EmailConfig) error {
	// encode the message with the template, the subject and the body
	msg := []byte(fmt.Sprintf(plainTemplate, e.Subject, e.Body))
	// create the auth object with the email credentials
	auth := smtp.PlainAuth("", conf.Address, conf.Password, conf.EmailHost)
	// create the server string with the host and the port and the receipts
	server := fmt.Sprintf("%s:%d", conf.EmailHost, conf.EmailPort)
	receipts := []string{e.To}
	// send the email
	var err error
	for i := 0; i < retries; i++ {
		if err = smtp.SendMail(server, auth, conf.Address, receipts, msg); err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}
