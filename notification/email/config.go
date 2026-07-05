package email

import (
	"net/mail"
	"strings"

	"github.com/simpleauthlink/authapi/notification"
)

// defaultSendRetries is the default number of retries to send the email.
const defaultSendRetries = 3

// EmailConfig struct represents the email configuration that is needed to send
// an email using and SMTP server. It includes the email address (used as the
// sender address but also as the username for the SMTP server), the email
// server hostname, its port and the password.
type EmailConfig struct {
	FromAddress  string `env:"EMAIL_ADDR" flag:"email-addr" usage:"account address that will send the email notifications"`
	SMTPUsername string `env:"EMAIL_USER" flag:"email-user" usage:"account username to get authorized into the SMTP server"`
	SMTPPassword string `env:"EMAIL_PASS" flag:"email-pass" usage:"account password to get authorized into the SMTP server"`
	SMTPServer   string `env:"EMAIL_HOST" flag:"email-host" usage:"SMTP server host"`
	SMTPPort     int    `env:"EMAIL_PORT" flag:"email-port" usage:"SMTP server port"`
	Retries      int
}

// Valid method checks if the email configuration is valid. It returns true if
// the sender name, the SMTP server and its port are not empty, and the sender
// email address is valid. It also sets the number of retries to the default
// value if it is not set.
func (cfg *EmailConfig) Valid() bool {
	if cfg.FromAddress == "" || cfg.SMTPServer == "" || cfg.SMTPPort == 0 {
		return false
	}
	fromAddress := strings.TrimSpace(cfg.FromAddress)
	fromAddress = strings.Trim(fromAddress, `"`)
	_, err := mail.ParseAddress(fromAddress)
	return err == nil
}

func (cfg *EmailConfig) Support(n notification.Notification) bool {
	_, ok := any(n).(*EmailNotification)
	return ok
}
