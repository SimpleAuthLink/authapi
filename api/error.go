package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var (
	// Decode data errors
	DecodeAppIDRequestErr       = newApiErr(1001, http.StatusBadRequest).With("could not decode app id request")
	DecodeTokenRequestErr       = newApiErr(1002, http.StatusBadRequest).With("could not decode token request")
	DecodeTokenStatusRequestErr = newApiErr(1003, http.StatusBadRequest).With("could not decode token status request")
	// Encode data errors
	EncodeAppIDResponseErr       = newApiErr(1010, http.StatusInternalServerError).With("could not encode app id response")
	EncodeTokenStatusResponseErr = newApiErr(1011, http.StatusInternalServerError).With("could not encode token status response")
	// Bad request errors
	InvalidAppHeadersErr     = newApiErr(1020, http.StatusBadRequest).With("invalid app headers")
	InvalidAppIDErr          = newApiErr(1021, http.StatusBadRequest).With("invalid app id")
	InvalidAppSecretErr      = newApiErr(1022, http.StatusBadRequest).With("invalid app secret")
	InvalidDemoEmailInboxErr = newApiErr(1023, http.StatusBadRequest).With("invalid demo email inbox")
	// Internal errors
	GenerateTokenErr = newApiErr(1030, http.StatusInternalServerError).With("could not generate token")
	GenerateEmailErr = newApiErr(1031, http.StatusInternalServerError).With("could not generate email")
	SendEmailErr     = newApiErr(1032, http.StatusInternalServerError).With("could not send email")
	InternalErr      = newApiErr(1033, http.StatusInternalServerError).With("internal server error")
)

type APIError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Err        string `json:"error,omitempty"`
	statusCode int
}

func (e *APIError) Bytes() []byte {
	bErr, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bErr
}

func (e *APIError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, error: %s, status_code: %d", e.Code, e.Message, e.Err, e.statusCode)
}

func (e *APIError) WithErr(err error) *APIError {
	if e.Err == "" {
		e.Err = err.Error()
		return e
	}
	e.Err = fmt.Sprintf("%s: %s", e.Err, err.Error())
	return e
}

func (e *APIError) With(msg string) *APIError {
	if e.Message == "" {
		e.Message = msg
		return e
	}
	e.Message = fmt.Sprintf("%s: %s", e.Message, msg)
	return e
}

func (e *APIError) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.statusCode)
	if _, err := w.Write(e.Bytes()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func newApiErr(code, status int) *APIError {
	return &APIError{
		Code:       code,
		statusCode: status,
	}
}
