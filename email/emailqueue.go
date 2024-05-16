package email

import (
	"context"
	"fmt"
	"net/smtp"
	"regexp"
	"sync"
	"time"
)

const (
	// plainTemplate is the template used to compose the email message. It
	// includes the subject and the body of the email.
	plainTemplate = "Subject: %s\r\n\r\n%s\r\n"
	// sendRetries is the number of retries to send the email.
	sendRetries = 3
)

// emailRgx is the regular expression used to validate an email address.
var emailRgx = regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,}$`)

// EmailConfig struct represents the email configuration that is needed to send
// an email using and SMTP server. It includes the email address (used as the
// sender address but also as the username for the SMTP server), the email
// server hostname, its port and the password.
type EmailConfig struct {
	Address       string
	EmailHost     string
	EmailPort     int
	Password      string
	DisposableSrc string
}

// Email struct represents the email that is going to be sent. It includes the
// recipient email address, the subject and the body of the email.
type Email struct {
	To      string
	Subject string
	Body    string
}

// EmailQueue struct represents the email queue. It includes the context and the
// cancel function to stop the queue, the configuration of the server to send
// the email, the list of emails to send, and the waiter to wait for the
// background process to finish.
type EmailQueue struct {
	ctx               context.Context
	cancel            context.CancelFunc
	cfg               *EmailConfig
	items             []*Email
	itemsMtx          sync.Mutex
	waiter            sync.WaitGroup
	disallowedDomains []string
}

// NewEmailQueue creates a new EmailQueue with the provided configuration.
func NewEmailQueue(ctx context.Context, cfg *EmailConfig) (*EmailQueue, error) {
	// check if the configuration is valid
	if cfg.Address == "" || !emailRgx.MatchString(cfg.Address) ||
		cfg.EmailHost == "" || cfg.EmailPort == 0 || cfg.Password == "" {
		return nil, ErrInvalidConfig
	}
	internalCtx, cancel := context.WithCancel(ctx)
	// load the disposable domains if a source is provided
	var err error
	disallowedDomains := []string{}
	if cfg.DisposableSrc != "" {
		disallowedDomains, err = LoadRemoteDisposableDomains(internalCtx, cfg.DisposableSrc)
	}
	// return the email queue
	return &EmailQueue{
		ctx:               internalCtx,
		cancel:            cancel,
		cfg:               cfg,
		items:             []*Email{},
		disallowedDomains: disallowedDomains,
	}, err
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
				e := eq.Pop()
				if e == nil {
					continue
				}
				if err := eq.Send(e); err != nil {
					fmt.Println(err)
				} else {
					eq.Pop()
				}
			}
			time.Sleep(time.Second)
		}
	}()
}

func (eq *EmailQueue) Stop() {
	eq.cancel()
	eq.waiter.Wait()
}

// Push method adds a new email to the queue.
func (eq *EmailQueue) Push(e *Email) error {
	// check if the email is valid
	if e.To == "" || !emailRgx.MatchString(e.To) || e.Subject == "" || e.Body == "" {
		return ErrInvalidEmail
	}
	// check if the email is allowed
	if !eq.Allowed(e.To) {
		return ErrDisallowedDomain
	}
	eq.itemsMtx.Lock()
	eq.items = append(eq.items, e)
	eq.itemsMtx.Unlock()
	return nil
}

// Top method returns the first email in the queue.
func (eq *EmailQueue) Top() *Email {
	eq.itemsMtx.Lock()
	defer eq.itemsMtx.Unlock()
	if len(eq.items) == 0 {
		return nil
	}
	return eq.items[0]
}

// Pop method removes the first email in the queue and returns it.
func (eq *EmailQueue) Pop() *Email {
	eq.itemsMtx.Lock()
	defer eq.itemsMtx.Unlock()
	if len(eq.items) == 0 {
		return nil
	}
	e := eq.items[0]
	eq.items = eq.items[1:]
	return e
}

// Send method sends the email using the queue configuration. It uses the
// email address as the sender address and the username for the SMTP server.
// It composes the email message, creates the auth object with the email
// credentials, the server string with the host and the port, and the receipts.
// Finally, it sends the email. If something fails during the process, it
// returns an error.
func (eq *EmailQueue) Send(e *Email) error {
	// check if the email is valid
	if e.To == "" || !emailRgx.MatchString(e.To) || e.Subject == "" || e.Body == "" {
		return ErrInvalidEmail
	}
	// check if the email is allowed
	if !eq.Allowed(e.To) {
		return ErrDisallowedDomain
	}
	// encode the message with the template, the subject and the body
	msg := []byte(fmt.Sprintf(plainTemplate, e.Subject, e.Body))
	// create the auth object with the email credentials
	auth := smtp.PlainAuth("", eq.cfg.Address, eq.cfg.Password, eq.cfg.EmailHost)
	// create the server string with the host and the port and the receipts
	server := fmt.Sprintf("%s:%d", eq.cfg.EmailHost, eq.cfg.EmailPort)
	receipts := []string{e.To}
	// send the email
	var err error
	for i := 0; i < sendRetries; i++ {
		if err = smtp.SendMail(server, auth, eq.cfg.Address, receipts, msg); err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}

// Allowed method checks if the email address is allowed. It compares the domain
// with a list of disallowed domains. It returns true if the email address is
// allowed, otherwise it returns false.
func (eq *EmailQueue) Allowed(address string) bool {
	if !emailRgx.MatchString(address) {
		return false
	}
	return CheckEmail(eq.disallowedDomains, address)
}
