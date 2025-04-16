package internal

import "fmt"

// Error represents an error with a message and a trace. It is a custom struct
// that can be used to wrap errors with additional information.
type Error struct {
	msg   string
	trace error
}

// NewErr creates a new Error instance with the given message. It is a helper
// function that simplifies the creation of Error instances in other packages.
func NewErr(msg string) *Error {
	return &Error{msg: msg}
}

// Error returns the error message. If the error has a trace, it is appended to
// the message. The error implements the error interface.
func (e *Error) Error() string {
	err := fmt.Errorf("%s", e.msg)
	if e.trace != nil {
		err = fmt.Errorf("%s: %w", err, e.trace)
	}
	return err.Error()
}

// With adds an error as a trace to the error. It is a helper function that
// simplifies the addition of traces to Error instances in other packages.
func (e *Error) With(err error) *Error {
	e.trace = err
	return e
}

// Withf adds a formatted error message as a trace to the error. It is a helper
// function that simplifies the addition of formatted traces to Error instances
// in other packages.
func (e *Error) Withf(tmpl string, args ...any) *Error {
	e.trace = fmt.Errorf(tmpl, args...)
	return e
}
