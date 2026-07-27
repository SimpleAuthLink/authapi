package email

import "go.k7z7z.cc/x/errors"

var (
	// ErrInvalidConfig is the error returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")
	// ErrInvalidEmail is the error returned when the email is invalid.
	ErrInvalidEmail = errors.New("invalid email")
	// ErrInvalidTemplate is the error returned when the template is invalid.
	ErrInvalidTemplate = errors.New("invalid template")
	// ErrSendEmail is the error returned when the email cannot be sent.
	ErrSendEmail = errors.New("error sending email")
	// ErrComposeEmail is the error returned when the email cannot be composed.
	ErrComposeEmail = errors.New("error composing email")
	// ErrCreatePart is the error returned when an email part cannot be written.
	ErrCreatePart = errors.New("error creating email part")
	// ErrParseAddress is the error returned when the email address cannot
	// be parsed.
	ErrParseAddress = errors.New("error parsing email address")
	// ErrSetBoundary is the error returned when the boundary cannot be set
	// when a multipart email is composed.
	ErrSetBoundary = errors.New("error setting boundary")
	// ErrWriteBody is the error returned when the email plain body cannot
	// be written.
	ErrWriteBody = errors.New("error writing email plain body")
	// ErrWriteHTMLBody is the error returned when the email HTML body cannot
	// be written.
	ErrWriteHTMLBody = errors.New("error writing email html body")
	// ErrCloseEmailWriter is the error returned when the email writer cannot
	// be closed after composing the email.
	ErrCloseEmailWriter = errors.New("error closing email writer")
)
