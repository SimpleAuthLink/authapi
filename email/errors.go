package email

import "fmt"

var (
	// ErrInvalidConfig is the error returned when the configuration is invalid.
	ErrInvalidConfig = fmt.Errorf("invalid configuration")
	// ErrInitQueue is the error returned when the queue cannot be initialized.
	ErrInitQueue = fmt.Errorf("error initializing the queue")
	// ErrInvalidDomain is the error returned when the domain is invalid.
	ErrInvalidDomain = fmt.Errorf("invalid domain")
	// ErrLoadingDisposableDomains is the error returned when the disposable
	// domains cannot be loaded.
	ErrLoadingDisposableDomains = fmt.Errorf("error loading disposable domains")
	// ErrDisallowedDomain is the error returned when the domain is disallowed.
	ErrDisallowedDomain = fmt.Errorf("disallowed domain")
	// ErrInvalidEmail is the error returned when the email is invalid.
	ErrInvalidEmail = fmt.Errorf("invalid email")
)
