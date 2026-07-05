package email

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/simpleauthlink/authapi/notification"
)

// EmailParams holds the recipient and subject for an email notification.
type EmailParams struct {
	To      string
	Subject string
}

// Valid reports whether the recipient is a parseable email address and the
// subject is non-empty.
func (p EmailParams) Valid() bool {
	_, err := mail.ParseAddress(p.To)
	return err == nil && p.Subject != ""
}

// EmailNotification is a notification that can be sent via SMTP. At least one
// of Body (HTML) or PlainBody (plain text) must be set.
type EmailNotification struct {
	Params    EmailParams
	Body      []byte
	PlainBody []byte
}

// Valid method checks if the email is valid. It returns true if the recipient
// email address, the subject and the body are not empty.
func (n *EmailNotification) Valid() bool {
	return n.Params.Valid() && max(len(n.Body), len(n.PlainBody)) > 0
}

// Send delivers the email via the SMTP server configured in the provided
// EmailConfig. It retries up to cfg.Retries times (or a default of 3) on
// transient failures.
func (n *EmailNotification) Send(conf notification.QueueConfig) error {
	cfg, ok := any(conf).(*EmailConfig)
	if !ok {
		return ErrInvalidConfig
	}
	// init SMTP auth
	var auth smtp.Auth
	if cfg.SMTPUsername != "" && cfg.SMTPPassword != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPServer)
	}
	// check if the email is valid
	if !n.Valid() {
		return ErrInvalidEmail
	}
	// compose the email body
	body, err := n.composeBody(cfg)
	if err != nil {
		return ErrComposeEmail.With(err)
	}
	receipts := []string{n.Params.To}
	// create the server string with the host and the port and the receipts
	server := fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort)
	// send the email
	retries := cfg.Retries
	if retries == 0 {
		retries = defaultSendRetries
	}
	for range retries {
		if err = smtp.SendMail(server, auth, cfg.FromAddress, receipts, body); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return ErrSendEmail.With(err)
	}
	return nil
}

func (n *EmailNotification) composeBody(conf *EmailConfig) ([]byte, error) {
	// parse 'to' email address
	to, err := mail.ParseAddress(n.Params.To)
	if err != nil {
		return nil, ErrParseAddress.With(err)
	}
	// create email headers
	var headers bytes.Buffer
	boundary := "----=_Part_0_123456789.123456789"
	_, _ = fmt.Fprintf(&headers, "From: %s\r\n", conf.FromAddress)
	_, _ = fmt.Fprintf(&headers, "To: %s\r\n", to.String())
	_, _ = fmt.Fprintf(&headers, "Subject: %s\r\n", n.Params.Subject)
	_, _ = fmt.Fprintf(&headers, "MIME-Version: 1.0\r\n")
	_, _ = fmt.Fprintf(&headers, "Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	_, _ = fmt.Fprintf(&headers, "\r\n") // blank line between headers and body
	// create multipart writer
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, ErrSetBoundary.With(err)
	}
	// plain text part
	if len(n.PlainBody) > 0 {
		textPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=\"UTF-8\""},
			"Content-Transfer-Encoding": {"7bit"},
		})
		if err != nil {
			return nil, ErrCreatePart.With(err)
		}
		if _, err := textPart.Write(n.PlainBody); err != nil {
			return nil, ErrWriteBody.With(err)
		}
	}
	// HTML part
	if len(n.Body) > 0 {
		htmlPart, err := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/html; charset=\"UTF-8\""},
			"Content-Transfer-Encoding": {"7bit"},
		})
		if err != nil {
			return nil, ErrCreatePart.With(err)
		}
		if _, err := htmlPart.Write(n.Body); err != nil {
			return nil, ErrWriteHTMLBody.With(err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, ErrCloseEmailWriter.With(err)
	}
	// combine headers and body and return the content
	var email bytes.Buffer
	email.Write(headers.Bytes())
	email.Write(body.Bytes())
	return email.Bytes(), nil
}
