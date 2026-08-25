package api

import (
	"net/http"
	"net/url"

	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/notification/email/templates/login"
	"github.com/simpleauthlink/authapi/token"
	"go.k7z7z.cc/x/net/http/io"
)

// generateAppIDHandler handles the request to generate an app id it decodes
// the app data from the request body and returns the app id in the response
// body. Every app data information is required to generate the app id. The
// app id is a self-contained representation of the app that can be used to
// generate tokens. It is created by encoding the app as a base64-encoded
// byte slice resulting in concatenating the app name, redirect uri, and
// session duration.
//
//	@Summary		Create an App
//	@Description	Register a new application by submitting its configuration
//	@Description	details, including the app name, session duration, redirect
//	@Description	URL, and secret. This process generates a unique AppID that
//	@Description	is used to identify your application and secure subsequent
//	@Description	API requests.
//	@Tags			apps
//	@Produce		json
//	@Param			request	body		api.AppIDRequest	true	"App ID Request"
//	@Success		200		{object}	api.AppIDResponse
//	@Failure		400		{object}	io.APIError
//	@Router			/apps [post]
func (s *APIService) generateAppIDHandler(w http.ResponseWriter, r *http.Request) {
	// decode the app data from the request body
	req := new(io.Request[AppIDRequest])
	if err := req.Read(r); err != nil {
		ErrDecodeAppIDRequest.WithErr(err).Write(w)
		return
	}
	// create the app from the data and check if it is valid
	app, err := req.Data.parseApp(s.cfg.Secret)
	if err != nil {
		ErrInvalidAppID.WithErr(err).Write(w)
		return
	}
	if err := app.Valid(app.AppSecretHash); err != nil {
		ErrInvalidAppID.WithErr(err).Write(w)
		return
	}
	// return the app id
	io.ResponseWith(&AppIDResponse{app.ID(app.Secret).String()}).WriteJSON(w)
}

// requestTokenHandler handles the request to request a new token for a
// given user email address. It takes the AppID and the AppSecret from the
// request headers, and the user email from the request body, encoded into
// a TokenRequest struct. It validates the input data, generates the user
// token with the user email address and send to it the resulting token.
//
//	@Summary		Request a new token for the user
//	@Description	Request an authentication token for a user by providing
//	@Description	their email address. The service generates a secure token
//	@Description	using cryptographic methods and sends it via email in a
//	@Description	magic link, enabling passwordless login. This endpoint
//	@Description	requires X-SIMPLEAUTHLINK-APPID and X-SIMPLEAUTHLINK-SECRET
//	@Description	auth headers to be defined.
//	@Tags			tokens
//	@Produce		json
//	@Security		X-SIMPLEAUTHLINK-APPID || X-SIMPLEAUTHLINK-SECRET
//	@Param			request	body	api.TokenRequest	true	"Token Request"
//	@Success		200
//	@Failure		400	{object}	io.APIError
//	@Failure		500	{object}	io.APIError
//	@Router			/tokens [post]
func (s *APIService) requestTokenHandler(w http.ResponseWriter, r *http.Request) {
	app, secret, err := appAndSecretFromRequest(r, []byte(s.cfg.Secret))
	if err != nil {
		err.Write(w)
		return
	}
	// AppID can not be nil because the appAndSecretFromRequest checks it, so
	// not nil-check is required here
	appID := app.ID(secret)
	// decode the token request from the request body
	req := new(io.Request[TokenRequest])
	if err := req.Read(r); err != nil {
		ErrDecodeTokenRequest.WithErr(err).Write(w)
		return
	}

	userEmail := new(token.Email).SetString(req.Data.Email)
	switch {
	case userEmail.Valid():
		// generate user genToken for the appID, with the secret and the given
		// user email. Nil-check is not required since at this point all
		// required information is valid and provided.
		genToken := appID.GenerateToken(*secret, *userEmail)

		// Parse error is not reachable at this point since the app comes
		// from a validated appID which requires a valid redirect URI
		linkURL, _ := url.Parse(app.RedirectURI)
		q := linkURL.Query()
		q.Set("token", genToken.String())
		q.Set("user", req.Data.Email)
		linkURL.RawQuery = q.Encode()

		// compose the email with the token
		loginData := login.Data{
			AppName: app.Name,
			Email:   req.Data.Email,
			Token:   genToken.String(),
			Link:    linkURL.String(),
		}
		loginEmail, err := login.Template.Compose(email.EmailParams{
			To:      req.Data.Email,
			Subject: loginData.Subject(),
		}, loginData)
		if err != nil {
			ErrGenerateNotification.WithErr(err).Write(w)
			return
		}
		// push the email to the notification queue
		if err := s.nq.Push(loginEmail); err != nil {
			ErrNotificationChannel.WithErr(err).Write(w)
			return
		}
	default:
		ErrInvalidNotificationChannel.Write(w)
		return
	}
	io.OkResponse().WriteJSON(w)
}

// verifyTokenHandler handles the a token verification request. It takes the
// AppID and the AppSecret from the request headers. It validates the token
// for the given app configuration and check the token expiration.
//
//	@Summary		Verify a user token
//	@Description	Validate a received token to ensure it is authentic,
//	@Description	unexpired, and properly associated with your application
//	@Description	and the user. This step is critical for maintaining secure
//	@Description	sessions. This endpoint requires X-SIMPLEAUTHLINK-APPID and
//	@Description	X-SIMPLEAUTHLINK-SECRET auth headers to be defined.
//	@Tags			tokens
//	@Produce		json
//	@Security		X-SIMPLEAUTHLINK-APPID || X-SIMPLEAUTHLINK-SECRET
//	@Param			request	body		api.TokenStatusRequest	true	"Token Request"
//	@Success		200		{object}	api.TokenStatusResponse
//	@Failure		400		{object}	io.APIError
//	@Router			/tokens [put]
func (s *APIService) verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	app, secret, err := appAndSecretFromRequest(r, []byte(s.cfg.Secret))
	if err != nil {
		err.Write(w)
		return
	}
	appID := app.ID(secret)
	// decode the token status request from the request body
	req := new(io.Request[TokenStatusRequest])
	if err := req.Read(r); err != nil {
		ErrDecodeTokenStatusRequest.WithErr(err).Write(w)
		return
	}
	// check if the token is valid
	tkn := new(token.Token).SetString(req.Data.Token)
	exp := tkn.Expiration().Time()
	io.ResponseWith(&TokenStatusResponse{
		Valid:      appID.VerifyToken(*tkn, *secret),
		Expiration: exp,
	}).WriteJSON(w)
}

// healthCheckHandler handles the ping request.
//
//	@Summary		Ping endpoint
//	@Description	This endpoint allows to a third party to check if the
//	@Description	service is alive and running.
//	@Success		200
//	@Router			/ping [get]
func (s *APIService) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	io.OkResponse().Write(w)
}
