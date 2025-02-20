package email

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"sync"
)

// defaultSendRetries is the default number of retries to send the email.
const defaultSendRetries = 3

// Email struct represents the email that is going to be sent. It includes the
// recipient email address, the subject and the body of the email.
type Email struct {
	To        string
	Subject   string
	Body      []byte
	PlainBody []byte
}

// Valid method checks if the email is valid. It returns true if the recipient
// email address, the subject and the body are not empty.
func (e *Email) Valid() bool {
	if e.Subject == "" || (len(e.Body) == 0 && len(e.PlainBody) == 0) {
		return false
	}
	_, err := mail.ParseAddress(e.To)
	return err == nil
}

// EmailConfig struct represents the email configuration that is needed to send
// an email using and SMTP server. It includes the email address (used as the
// sender address but also as the username for the SMTP server), the email
// server hostname, its port and the password.
type EmailConfig struct {
	FromName     string
	FromAddress  string
	SMTPUsername string
	SMTPPassword string
	SMTPServer   string
	SMTPPort     int
	Retries      int
	ErrorCh      chan error
}

// Valid method checks if the email configuration is valid. It returns true if
// the sender name, the SMTP server and its port are not empty, and the sender
// email address is valid. It also sets the number of retries to the default
// value if it is not set.
func (cfg *EmailConfig) Valid() bool {
	if cfg.FromName == "" || cfg.SMTPServer == "" || cfg.SMTPPort == 0 {
		return false
	}
	if cfg.Retries == 0 {
		cfg.Retries = defaultSendRetries
	}
	_, err := mail.ParseAddress(cfg.FromAddress)
	return err == nil
}

// EmailQueue struct represents the email queue. It includes the context and the
// cancel function to stop the queue, the configuration of the server to send
// the email, the list of emails to send, and the waiter to wait for the
// background process to finish.
type EmailQueue struct {
	ctx      context.Context
	cancel   context.CancelFunc
	cfg      *EmailConfig
	auth     smtp.Auth
	items    []*Email
	itemsMtx sync.Mutex
	waiter   sync.WaitGroup
	errCh    chan error
}

// NewEmailQueue creates a new EmailQueue with the provided configuration.
func NewEmailQueue(ctx context.Context, cfg *EmailConfig) (*EmailQueue, error) {
	// check if the configuration is valid
	if !cfg.Valid() {
		return nil, ErrInvalidConfig
	}
	// init the email queue
	internalCtx, cancel := context.WithCancel(ctx)
	eq := &EmailQueue{
		ctx:    internalCtx,
		cancel: cancel,
		cfg:    cfg,
		items:  []*Email{},
		errCh:  cfg.ErrorCh,
	}
	// init SMTP auth
	if cfg.SMTPUsername != "" && cfg.SMTPPassword != "" {
		eq.auth = smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPServer)
	}
	// return the email queue
	return eq, nil
}

// Start method starts the email queue. It listens for new emails in the queue
// and sends them using the provided configuration.
func (eq *EmailQueue) Start() {
	eq.waiter.Add(1)
	go func() {
		defer eq.waiter.Done()
		for {
			select {
			case <-eq.ctx.Done():
				return
			default:
				e, ok := eq.Pop()
				if !ok {
					continue
				}
				if err := eq.Send(e); err != nil {
					if eq.errCh != nil {
						eq.errCh <- err
					}
				}
			}
		}
	}()
}

// Stop method stops the email queue.
func (eq *EmailQueue) Stop() {
	eq.cancel()
	eq.waiter.Wait()
}

// Push method adds a new email to the queue.
func (eq *EmailQueue) Push(e Email) error {
	// check if the email is valid
	if !e.Valid() {
		return ErrInvalidEmail
	}
	eq.itemsMtx.Lock()
	eq.items = append(eq.items, &e)
	eq.itemsMtx.Unlock()
	return nil
}

// Pop method removes the first email in the queue and returns it.
func (eq *EmailQueue) Pop() (Email, bool) {
	eq.itemsMtx.Lock()
	defer eq.itemsMtx.Unlock()
	if len(eq.items) == 0 {
		return Email{}, false
	}
	e := eq.items[0]
	eq.items = eq.items[1:]
	return *e, true
}

// Send method sends the email using the queue configuration. It uses the
// email address as the sender address and the username for the SMTP server.
// It composes the email message, creates the auth object with the email
// credentials, the server string with the host and the port, and the receipts.
// Finally, it sends the email. If something fails during the process, it
// returns an error. It can be used even the queue is not started.
func (eq *EmailQueue) Send(e Email) error {
	// check if the email is valid
	if !e.Valid() {
		return ErrInvalidEmail
	}
	// compose the email body
	body, err := eq.composeBody(e)
	if err != nil {
		return ErrComposeEmail.With(err)
	}
	// create the server string with the host and the port and the receipts
	server := fmt.Sprintf("%s:%d", eq.cfg.SMTPServer, eq.cfg.SMTPPort)
	receipts := []string{e.To}
	// send the email
	for i := 0; i < eq.cfg.Retries; i++ {
		if err = smtp.SendMail(server, eq.auth, eq.cfg.FromAddress, receipts, body); err == nil {
			break
		}
	}
	if err != nil {
		return ErrSendEmail.With(err)
	}
	return nil
}

// composeBody creates the email body with the message data. It creates a
// multipart email with a plain text and an HTML part. It returns the email
// content as a byte slice or an error if the body could not be composed.
func (eq *EmailQueue) composeBody(msg Email) ([]byte, error) {
	// parse 'to' email address
	to, err := mail.ParseAddress(msg.To)
	if err != nil {
		return nil, ErrParseAddress.With(err)
	}
	// create email headers
	var headers bytes.Buffer
	boundary := "----=_Part_0_123456789.123456789"
	headers.WriteString(fmt.Sprintf("From: %s\r\n", eq.cfg.FromAddress))
	headers.WriteString(fmt.Sprintf("To: %s\r\n", to.String()))
	headers.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	headers.WriteString("MIME-Version: 1.0\r\n")
	headers.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	headers.WriteString("\r\n") // blank line between headers and body
	// create multipart writer
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.SetBoundary(boundary); err != nil {
		return nil, ErrSetBoundary.With(err)
	}
	// plain text part
	textPart, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/plain; charset=\"UTF-8\""},
		"Content-Transfer-Encoding": {"7bit"},
	})
	if _, err := textPart.Write(msg.PlainBody); err != nil {
		return nil, ErrWriteBody.With(err)
	}
	// HTML part
	htmlPart, _ := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=\"UTF-8\""},
		"Content-Transfer-Encoding": {"7bit"},
	})
	if _, err := htmlPart.Write(msg.Body); err != nil {
		return nil, ErrWriteHTMLBody.With(err)
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
