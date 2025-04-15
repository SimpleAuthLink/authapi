package io

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an error response from the API. It includes a code,
// message, and optional error string. The StatusCode field is used to set the
// HTTP status code for the response. It has some methods to manipulate the
// error message and write the error response to an http.ResponseWriter.
type APIError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Err        string `json:"error,omitempty"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface for APIError. It returns a string
// representation of the error, including the code, message, error string,
// and status code. It can be used for logging or debugging purposes.
func (e *APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, error: %s, status_code: %d", e.Code, e.Message, e.Err, e.StatusCode)
}

// WithErr appends an error message to the existing error string in the
// APIError. If the existing error string is empty, it sets it to the new
// error message. This method is useful for chaining error messages together
// for better debugging and logging. It returns the updated APIError instance
// and also updates the current instance.
func (e *APIError) WithErr(err error) *APIError {
	if e.Err == "" {
		e.Err = err.Error()
		return e
	}
	e.Err = fmt.Sprintf("%s: %s", e.Err, err.Error())
	return e
}

// With appends a string message to the existing message in the APIError. If
// the existing message is empty, it sets it to the new message. This method
// is useful for chaining messages together for better debugging and logging.
// It returns the updated APIError instance and also updates the current
// instance.
func (e *APIError) With(msg string) *APIError {
	if e.Message == "" {
		e.Message = msg
		return e
	}
	e.Message = fmt.Sprintf("%s: %s", e.Message, msg)
	return e
}

// WriteJSON writes the APIError as a JSON response to the provided
// http.ResponseWriter. It sets the Content-Type header to "application/json"
// and writes the status code and serialized error bytes to the response.
// If an error occurs during writing, it writes an internal server error
// response instead.
func (e *APIError) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	if _, err := w.Write(e.bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// bytes serializes the APIError to JSON bytes. If an error occurs during
// serialization, it returns nil.
func (e *APIError) bytes() []byte {
	bErr, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bErr
}

// NewAPIError creates a new APIError instance with the provided code and
// status code. It initializes the error string and message to empty strings.
// This function is useful for creating a new APIError instance with
// specific error codes and status codes. It returns a pointer to the
// newly created APIError instance.
func NewAPIError(code, status int) *APIError {
	return &APIError{
		Code:       code,
		StatusCode: status,
	}
}
