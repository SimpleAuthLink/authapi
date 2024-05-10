package authapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	APP_SECRET_HEADER = "APP_SECRET"
	USER_TOKEN_QUERY  = "token"
)

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
	// generate token
	magicLink, err := s.magicLink(appSecret, req.Email)
	if err != nil {
		log.Println("ERR: error generating token:", err)
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	// compose and send email
	email := &Email{
		To:      req.Email,
		Subject: userTokenSubject,
		Body:    fmt.Sprintf(userTokenBody, magicLink),
	}
	if err := email.Send(&s.cfg.EmailConfig); err != nil {
		log.Println("ERR: error sending email:", err)
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

func (s *Service) validateUserTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get(USER_TOKEN_QUERY)
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}
	if !s.validUserToken(token) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	if _, err := w.Write([]byte("Ok")); err != nil {
		log.Println("ERR: error sending response:", err)
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}

func (s *Service) appTokenHandler(w http.ResponseWriter, r *http.Request) {
	// read body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
	app := &AppRequest{}
	if err := json.Unmarshal(body, app); err != nil {
		http.Error(w, "error parsing request body", http.StatusBadRequest)
		return
	}
	// generate token
	appId, secret, err := s.authApp(app.Name, app.Email, app.Callback, app.Duration)
	if err != nil {
		log.Println("ERR: error generating token:", err)
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	// send email
	email := &Email{
		To:      app.Email,
		Subject: appTokenSubject,
		Body:    fmt.Sprintf(appTokenBody, app.Name, appId, secret),
	}
	if err := email.Send(&s.cfg.EmailConfig); err != nil {
		log.Println("ERR: error sending email:", err)
		http.Error(w, "error sending email", http.StatusInternalServerError)
		return
	}
	// send response
	if _, err := w.Write([]byte("Ok")); err != nil {
		http.Error(w, "error sending response", http.StatusInternalServerError)
		return
	}
}
