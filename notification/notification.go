package notification

import "net/mail"

type NotificationParams struct {
	To      string
	Subject string
}

func (p NotificationParams) Valid() bool {
	_, err := mail.ParseAddress(p.To)
	return err == nil && p.Subject != ""
}

type Notification struct {
	Params    NotificationParams
	Body      []byte
	PlainBody []byte
}

// Valid method checks if the email is valid. It returns true if the recipient
// email address, the subject and the body are not empty.
func (n *Notification) Valid() bool {
	return n.Params.Valid() && max(len(n.Body), len(n.PlainBody)) > 0
}

type Queue interface {
	Start()
	Stop()
	Pop() (Notification, bool)
	Push(Notification) error
	Send(Notification) error
}
