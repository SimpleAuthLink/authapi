package authapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/simpleauthlink/authapi/db"
	"github.com/simpleauthlink/authapi/email"
)

const (
	// APP_SECRET_HEADER is the http header key for the app secret
	APP_SECRET_HEADER = "APP_SECRET"
	// USER_TOKEN_QUERY is the url query key for the user token
	USER_TOKEN_QUERY = "token"
)

// userTokenHandler method generates a token for the user and sends it via email
// to the user's email address. The token is generated based on the app id
// and the user's email address. The token is stored in the database with an
// expiration time. It gets the app secret from the APP_SECRET_HEADER header
// and the user's email address from the request body. If it success it sends
// an "Ok" response. If something goes wrong, it sends an internal server error
// response. If the app secret is missing or the request body is invalid, it
// sends a bad request response.
func (s *Service) userTokenHandler(w http.ResponseWriter, r *http.Request) {
	// read the app token header
	appSecret := r.Header.Get(APP_SECRET_HEADER)
	if appSecret == "" {
		http.Error(w, "missing app token", http.StatusBadRequest)
		return
	}
	// read body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("ERR: error reading request body:", err)
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
	// parse request
	req := &TokenRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		log.Println("ERR: error parsing request body:", err)
		http.Error(w, "error parsing request body", http.StatusBadRequest)
		return
	}
	// check if the email is allowed
	if !s.emailQueue.Allowed(req.Email) {
		http.Error(w, "disallowed domain", http.StatusBadRequest)
		return
	}
	// generate token
	magicLink, token, err := s.magicLink(appSecret, req.Email)
	if err != nil {
		log.Println("ERR: error generating token:", err)
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	// compose and push the email to the queue to be sent, if it fails, delete
	// the token from the database, log the error and send an error response
	if err := s.emailQueue.Push(&email.Email{
		To:      req.Email,
		Subject: userTokenSubject,
		Body:    fmt.Sprintf(userTokenBody, magicLink),
	}); err != nil {
		log.Println("ERR: error sending email:", err)
		if err := s.db.DeleteToken(db.Token(token)); err != nil {
			log.Println("ERR: error deleting token:", err)
		}
		http.Error(w, "error sending email", http.StatusInternalServerError)
		return
	}
	// send response
	if _, err := w.Write([]byte("Ok")); err != nil {
		log.Println("ERR: error sending response:", err)
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}

// validateUserTokenHandler method validates the user token. It gets the token
// from the USER_TOKEN_QUERY query string and checks if it is valid. If the
// token is valid, it sends a response with the "Ok" message. If the token is
// invalid, it sends an unauthorized response. If the token is missing, it
// sends a bad request response.
func (s *Service) validateUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	// read the app token header
	appSecret := r.Header.Get(APP_SECRET_HEADER)
	if appSecret == "" {
		http.Error(w, "missing app token", http.StatusBadRequest)
		return
	}
	// get the token from the query
	token := r.URL.Query().Get(USER_TOKEN_QUERY)
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	// validate the token
	if !s.validUserToken(token, appSecret) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if _, err := w.Write([]byte("Ok")); err != nil {
		log.Println("ERR: error sending response:", err)
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}

// appTokenHandler method generates creates an app in the service, it generates
// an app id and a secret for the app. It sends the app id and the secret via
// email to the app's email address. It gets the app name, email, callback, and
// duration from the request body. If it success it sends an "Ok" response. If
// something goes wrong, it sends an internal server error response. If the
// request body is invalid, it sends a bad request response.
func (s *Service) appTokenHandler(w http.ResponseWriter, r *http.Request) {
	// read body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("ERR: error reading request body:", err)
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
	app := &AppRequest{}
	if err := json.Unmarshal(body, app); err != nil {
		log.Println("ERR: error parsing request body:", err)
		http.Error(w, "error parsing request body", http.StatusBadRequest)
		return
	}
	// check if the email is allowed
	if !s.emailQueue.Allowed(app.Email) {
		http.Error(w, "disallowed domain", http.StatusBadRequest)
		return
	}
	// generate token
	appId, secret, err := s.authApp(app.Name, app.Email, app.RedirectURL, app.Duration)
	if err != nil {
		log.Println("ERR: error generating token:", err)
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	// compose and push the email to the queue to be sent if it fails, delete
	// the app from the database, log the error and send an error response
	if err := s.emailQueue.Push(&email.Email{
		To:      app.Email,
		Subject: appTokenSubject,
		Body:    fmt.Sprintf(appTokenBody, app.Name, appId, secret),
	}); err != nil {
		log.Println("ERR: error sending email:", err)
		if err := s.removeApp(appId); err != nil {
			log.Println("ERR: error deleting app:", err)
		}
		http.Error(w, "error sending email", http.StatusInternalServerError)
		return
	}
	// send response
	if _, err := w.Write([]byte("Ok")); err != nil {
		log.Println("ERR: error sending response:", err)
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}

// updateAppHandler method updates an app in the service. It gets the app id
// from the URL path and the app name, callback, and duration from the request
// body. If the app id is missing, it sends a bad request response. If the app
// is not found, it sends a not found response. If it success it sends an Ok
// response. If something goes wrong, it sends an internal server error
// response.
func (s *Service) updateAppHandler(w http.ResponseWriter, r *http.Request) {
	// read the app token header
	appSecret := r.Header.Get(APP_SECRET_HEADER)
	if appSecret == "" {
		http.Error(w, "missing app token", http.StatusBadRequest)
		return
	}
	// get the token from the query
	token := r.URL.Query().Get(USER_TOKEN_QUERY)
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	// validate the token and get the app id
	appId, valid := s.validAdminToken(token, appSecret)
	if !valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	// read body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("ERR: error reading request body:", err)
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
	// decode the app from the request
	app := &AppRequest{}
	if err := json.Unmarshal(body, app); err != nil {
		log.Println("ERR: error parsing request body:", err)
		http.Error(w, "error parsing request body", http.StatusBadRequest)
		return
	}
	// update the app in the database
	if err := s.updateAppMetadata(appId, app.Name, app.RedirectURL, app.Duration); err != nil {
		log.Println("ERR: error updating app:", err)
		http.Error(w, "error updating app", http.StatusInternalServerError)
		return
	}
	// send response
	if _, err := w.Write([]byte("Ok")); err != nil {
		log.Println("ERR: error sending response:", err)
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}

// delAppHandler method deletes an app from the service. It gets the app id from
// the token provided in the URL query. If the token is missing, it sends a bad
// request response. If the token is invalid or is not an admin token, it sends
// an unauthorized response. If it success it sends an Ok response. If something
// goes wrong, it sends an internal server error response.
func (s *Service) delAppHandler(w http.ResponseWriter, r *http.Request) {
	// read the app token header
	appSecret := r.Header.Get(APP_SECRET_HEADER)
	if appSecret == "" {
		http.Error(w, "missing app token", http.StatusBadRequest)
		return
	}
	// get the token from the query
	token := r.URL.Query().Get(USER_TOKEN_QUERY)
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	// validate the token and get the app id
	appId, valid := s.validAdminToken(token, appSecret)
	if !valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	// remove the app from the service
	if err := s.removeApp(appId); err != nil {
		log.Println("ERR: error deleting app:", err)
		http.Error(w, "error deleting app", http.StatusInternalServerError)
		return
	}
	// send response
	if _, err := w.Write([]byte("Ok")); err != nil {
		log.Println("ERR: error sending response:", err)
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}
