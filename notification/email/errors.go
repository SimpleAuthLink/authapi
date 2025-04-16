package email

import "github.com/simpleauthlink/authapi/internal"

var (
	// ErrInvalidConfig is the error returned when the configuration is invalid.
	ErrInvalidConfig = internal.NewErr("invalid configuration")
	// ErrInitQueue is the error returned when the queue cannot be initialized.
	ErrInitQueue = internal.NewErr("error initializing the queue")
	// ErrInvalidEmail is the error returned when the email is invalid.
	ErrInvalidEmail = internal.NewErr("invalid email")
	// ErrInvalidTemplate is the error returned when the template is invalid.
	ErrInvalidTemplate = internal.NewErr("invalid template")
	// ErrSendEmail is the error returned when the email cannot be sent.
	ErrSendEmail = internal.NewErr("error sending email")
	// ErrComposeEmail is the error returned when the email cannot be composed.
	ErrComposeEmail = internal.NewErr("error composing email")
	// ErrParseAddress is the error returned when the email address cannot
	// be parsed.
	ErrParseAddress = internal.NewErr("error parsing email address")
	// ErrSetBoundary is the error returned when the boundary cannot be set
	// when a multipart email is composed.
	ErrSetBoundary = internal.NewErr("error setting boundary")
	// ErrWriteHTMLBody is the error returned when the email plain body cannot
	// be written.
	ErrWriteBody = internal.NewErr("error writing email plain body")
	// ErrWriteHTMLBody is the error returned when the email HTML body cannot
	// be written.
	ErrWriteHTMLBody = internal.NewErr("error writing email html body")
	// ErrCloseEmailWriter is the error returned when the email writer cannot
	// be closed after composing the email.
	ErrCloseEmailWriter = internal.NewErr("error closing email writer")
)
