package api

import (
	"net/http"

	"go.k7z7z.cc/x/net/http/io"
)

var (
	// Decode data errors
	ErrDecodeAppIDRequest       = io.NewAPIError(4001, http.StatusBadRequest).With("could not decode app id request")
	ErrDecodeTokenRequest       = io.NewAPIError(4002, http.StatusBadRequest).With("could not decode token request")
	ErrDecodeTokenStatusRequest = io.NewAPIError(4003, http.StatusBadRequest).With("could not decode token status request")
	// Bad request errors
	ErrInvalidAppHeaders          = io.NewAPIError(4004, http.StatusBadRequest).With("invalid app headers")
	ErrInvalidAppID               = io.NewAPIError(4005, http.StatusBadRequest).With("invalid app id")
	ErrInvalidAppSecret           = io.NewAPIError(4006, http.StatusBadRequest).With("invalid app secret")
	ErrInvalidNotificationChannel = io.NewAPIError(4007, http.StatusBadRequest).With("invalid notification channel identified")
	// Internal errors
	ErrGenerateToken        = io.NewAPIError(5001, http.StatusInternalServerError).With("could not generate token")
	ErrGenerateNotification = io.NewAPIError(5002, http.StatusInternalServerError).With("could not generate notification")
	ErrNotificationChannel  = io.NewAPIError(5003, http.StatusInternalServerError).With("could not send notification")
)
