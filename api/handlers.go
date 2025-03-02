package api

import (
	"net/http"

	"github.com/simpleauthlink/authapi/notification"
	"github.com/simpleauthlink/authapi/notification/templates/login"
	"github.com/simpleauthlink/authapi/token"
)

func (s *Service) generateAppIDHandler(w http.ResponseWriter, r *http.Request) {
	// decode the app data from the request body
	req := new(Request[AppIDRequest])
	if err := req.Read(r); err != nil {
		DecodeAppIDRequestErr.WithErr(err).Write(w)
		return
	}
	// create the app from the data and check if it is valid
	app := req.Data.parseApp()
	if !app.Valid() {
		InvalidAppIDErr.Write(w)
		return
	}
	// return the app id
	if err := ResponseWith(&AppIDResponse{app.ID().String()}).Write(w); err != nil {
		EncodeAppIDResponseErr.WithErr(err).Write(w)
	}
}

func (s *Service) requestTokenHandler(w http.ResponseWriter, r *http.Request) {
	// get the app id from the request header
	strAppID, strAppSecret, err := appConfigFromRequest(r)
	if err != nil {
		InvalidAppHeadersErr.WithErr(err).Write(w)
		return
	}
	// decode the app id get the app from it
	appID := new(token.AppID).SetString(strAppID)
	app := new(token.App).SetID(appID)
	// check if the app id is valid (it should be a valid app)
	if !app.Valid() {
		InvalidAppIDErr.Write(w)
		return
	}
	// decode the token request from the request body
	req := new(Request[TokenRequest])
	if err := req.Read(r); err != nil {
		DecodeTokenRequestErr.WithErr(err).Write(w)
		return
	}
	// generate user token
	secret := new(token.Secret).SetParts([]byte(s.cfg.Secret), []byte(strAppSecret))
	if !secret.Valid() {
		InvalidAppSecretErr.Write(w)
		return
	}
	token := appID.GenerateToken(*secret, req.Data.Email)
	if token == nil {
		GenerateTokenErr.With(req.Data.Email).Write(w)
		return
	}
	// compose the email with the token
	loginData := login.Data{
		AppName: app.Name,
		Email:   req.Data.Email,
		Token:   token.String(),
		Link:    app.RedirectURI + token.String(),
	}
	loginEmail, err := login.Template.Compose(notification.NotificationParams{
		To:      req.Data.Email,
		Subject: loginData.Subject(),
	}, loginData)
	if err != nil {
		GenerateEmailErr.WithErr(err).Write(w)
		return
	}
	// push the email to the notification queue
	if err := s.nq.Push(loginEmail); err != nil {
		SendEmailErr.WithErr(err).Write(w)
		return
	}
	if err := OkResponse().Write(w); err != nil {
		InternalErr.WithErr(err).Write(w)
	}
}

func (s *Service) verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	// get the app id from the request header
	strAppID, strAppSecret, err := appConfigFromRequest(r)
	if err != nil {
		InvalidAppHeadersErr.WithErr(err).Write(w)
		return
	}
	// decode the app id get the app from it
	appID := new(token.AppID).SetString(strAppID)
	app := new(token.App).SetID(appID)
	// check if the app id is valid (it should be a valid app)
	if !app.Valid() {
		InvalidAppIDErr.Write(w)
		return
	}
	// decode the token status request from the request body
	req := new(Request[TokenStatusRequest])
	if err := req.Read(r); err != nil {
		DecodeTokenStatusRequestErr.WithErr(err).Write(w)
		return
	}
	// check if the token is valid
	tkn := new(token.Token).SetString(req.Data.Token)
	exp := tkn.Expiration().Time()
	secret := new(token.Secret).SetParts([]byte(s.cfg.Secret), []byte(strAppSecret))
	if !secret.Valid() {
		InvalidAppSecretErr.Write(w)
		return
	}
	ok := appID.VerifyToken(*tkn, *secret, req.Data.Email)
	if err := ResponseWith(&TokenStatusResponse{
		Valid:      ok,
		Expiration: exp,
	}).Write(w); err != nil {
		EncodeTokenStatusResponseErr.WithErr(err).Write(w)
		return
	}
}
