package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/simpleauthlink/authapi/api"
	"github.com/simpleauthlink/authapi/notification"
	"github.com/simpleauthlink/authapi/notification/email"
	"github.com/simpleauthlink/authapi/notification/email/templates/login"
	"go.k7z7z.cc/x/flag"
	"go.k7z7z.cc/x/log"
	"go.k7z7z.cc/x/net/http/io"
	"go.k7z7z.cc/x/net/smtp/testsmtp"
	"go.k7z7z.cc/x/proc"
)

type DemoConfig struct {
	APIHost string `env:"HOST" flag:"host" sflag:"h" usage:"service host"`
	APIPort int    `env:"PORT" flag:"port" sflag:"p" usage:"service port"`
	Secret  string `env:"SECRET" flag:"secret" sflag:"s" usage:"secret used to generate the tokens"`
}

var demoMailInbox = make(chan string, 1)

func main() {
	log.SetLevel(log.DebugLevel)
	c := &DemoConfig{
		APIHost: "0.0.0.0",
		APIPort: 8080,
		Secret:  "demo-service",
	}
	if err := flag.UnmarshalArgs(os.Args[1:], c); err != nil {
		log.Errw("error parsing flags", "err", err)
		return
	}
	log.Infow("config loaded", "config", c)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	demoMailServer := testsmtp.NewServer(c.APIHost, 2525, demoMailInbox)
	if err := demoMailServer.Start(ctx); err != nil {
		log.Errw("error starting test smtp server", "err", err)
		return
	}
	emailConfig := &email.EmailConfig{
		FromAddress: "demo@simpleauth.link",
		SMTPServer:  c.APIHost,
		SMTPPort:    2525,
	}
	// create the email queue
	notificationQueue, err := notification.NewQueue(ctx, 10, emailConfig)
	if err != nil {
		log.Errw("error initializing notification queue", "err", err)
		return
	}
	// start the email queue and defer to stop it
	notificationQueue.Start(5)
	defer notificationQueue.Stop()
	// create the apiService
	apiService, err := api.New(ctx, &api.Config{
		Server:     c.APIHost,
		ServerPort: c.APIPort,
		Secret:     c.Secret,
	}, notificationQueue)
	if err != nil {
		log.Errw("error creating api service", "err", err)
		return
	}
	if err := apiService.Get("/demo/inbox", demoInboxHandler); err != nil {
		log.Errw("error adding new handler for inbox", "err", err)
		return
	}
	// start the service and wait for it to finish
	if err = new(proc.Shutdown).WithInterrupt().WithRecovery().WithRun(func(ctx context.Context) error {
		return apiService.Start()
	}).WithCallback(func(_ context.Context) error {
		return apiService.Stop()
	}).Wait(ctx); err != nil {
		log.Errw("error during shutdown", "err", err)
		return
	}
	log.Info("HTTP server stopped, exiting...")
}

// demoInboxHandler retrieves the latest token for a demo email.
//
//	@Summary		Demo email inbox
//	@Description	Retrieves the latest token for an email sent to a demo
//	@Description	inbox. The stream closes once the token is delivered.
//	@Description	Only available in Demo environments.
//	@Tags			default
//	@Produce		text/event-stream
//	@Param			email	query		string	true	"Email address to get the token for"
//	@Success		200		{string}	string	"Server-Sent Events stream; the token arrives in the data field"
//	@Failure		400		{object}	io.APIError
//	@Router			/demo/inbox [get]
func demoInboxHandler(w http.ResponseWriter, r *http.Request) {
	// get the email from get parameters
	email := r.URL.Query().Get("email")
	if email == "" {
		io.NewAPIError(400001, http.StatusBadRequest).Write(w)
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
		case <-clientGone:
			return
		case msg := <-demoMailInbox:
			// find the token in the email
			if testToken := login.FindToken(email, msg); testToken != nil {
				// send an event to the client with the token in the "data" field
				if _, err := fmt.Fprintf(w, "data: %s\n\n", testToken); err != nil {
					return
				}
				if err := rc.Flush(); err != nil {
					return
				}
				// close the stream once the token is delivered so that the
				// response completes and clients (like swagger-ui) can show it
				return
			}
		}
	}
}
