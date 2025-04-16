package api

import (
	"net/http"

	"github.com/simpleauthlink/authapi/api/io"
)

var (
	// Decode data errors
	DecodeAppIDRequestErr       = io.NewAPIError(1001, http.StatusBadRequest).With("could not decode app id request")
	DecodeTokenRequestErr       = io.NewAPIError(1002, http.StatusBadRequest).With("could not decode token request")
	DecodeTokenStatusRequestErr = io.NewAPIError(1003, http.StatusBadRequest).With("could not decode token status request")
	// Encode data errors
	EncodeAppIDResponseErr       = io.NewAPIError(1010, http.StatusInternalServerError).With("could not encode app id response")
	EncodeTokenStatusResponseErr = io.NewAPIError(1011, http.StatusInternalServerError).With("could not encode token status response")
	// Bad request errors
	InvalidAppHeadersErr     = io.NewAPIError(1020, http.StatusBadRequest).With("invalid app headers")
	InvalidAppIDErr          = io.NewAPIError(1021, http.StatusBadRequest).With("invalid app id")
	InvalidAppSecretErr      = io.NewAPIError(1022, http.StatusBadRequest).With("invalid app secret")
	InvalidDemoEmailInboxErr = io.NewAPIError(1023, http.StatusBadRequest).With("invalid demo email inbox")
	// Internal errors
	GenerateTokenErr = io.NewAPIError(1030, http.StatusInternalServerError).With("could not generate token")
	GenerateEmailErr = io.NewAPIError(1031, http.StatusInternalServerError).With("could not generate email")
	SendEmailErr     = io.NewAPIError(1032, http.StatusInternalServerError).With("could not send email")
	InternalErr      = io.NewAPIError(1033, http.StatusInternalServerError).With("internal server error")
)
