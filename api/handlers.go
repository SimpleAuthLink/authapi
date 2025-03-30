package api

import (
	"fmt"
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
	ResponseWith(&AppIDResponse{app.ID().String()}).WriteJSON(w)
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
	OkResponse().WriteJSON(w)
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
	ResponseWith(&TokenStatusResponse{
		Valid:      ok,
		Expiration: exp,
	}).WriteJSON(w)
}

func (s *Service) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	OkResponse().Write(w)
}

func (s *Service) demoInboxHandler(w http.ResponseWriter, r *http.Request) {
	// get the email from get parameters
	email := r.URL.Query().Get("email")
	if email == "" {
		InvalidDemoEmailInboxErr.Write(w)
		return
	}
	// set http headers required for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// create a channel for client disconnection
	clientGone := r.Context().Done()
	// create a response controller
	rc := http.NewResponseController(w)
	for {
		select {
		case <-s.ctx.Done():
		case <-clientGone:
			return
		case msg := <-s.demoMailInbox:
			// find the token in the email
			if testToken := login.FindToken(email, msg); testToken != nil {
				// send an event to the client with the token in the "data" field
				if _, err := fmt.Fprintf(w, "data: %s\n\n", testToken); err != nil {
					return
				}
				if err := rc.Flush(); err != nil {
					return
				}
			}
		}
	}
}
