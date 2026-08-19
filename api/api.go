// Package api provides the HTTP API for the SimpleAuth.link service.
//
//	@title						SimpleAuth.link API
//	@version					1.0
//	@description				Passwordless authentication API for your users using just an email address.
//
//	@contact.name				API Support
//	@contact.url				https://simpleauth.link
//	@contact.email				info@simpleauth.link
//
//	@license.name				AGPL-3.0
//	@license.url				https://github.com/SimpleAuthLink/authapi/blob/main/LICENSE
//
//	@host						api.simpleauth.link
//	@BasePath					/
//	@schemes					https
//
//	@securityDefinitions.apikey	AppID
//	@in							header
//	@name						AppID
//
//	@securityDefinitions.apikey	AppSecret
//	@in							header
//	@name						AppSecret
//
//	@tag.name					apps
//	@tag.description			Create and manage your App
//	@tag.docs.url				https://docs.simpleauth.link/about/apps
//	@tag.docs.description		About Apps
//
//	@tag.name					tokens
//	@tag.description			Request and verify tokens
//	@tag.docs.url				https://docs.simpleauth.link/about/tokens
//	@tag.docs.description		About Tokens
package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/simpleauthlink/authapi/notification"
	xhttp "go.k7z7z.cc/x/net/http"
)

// Config struct represents the configuration for the API service. It contains
// the server address, server port, and secret key for the service. The server
// address is the address where the service will listen for incoming requests,
// and the server port is the port number where the service will listen for
// incoming requests. The secret key is used to sign and verify tokens.
type Config struct {
	Server     string
	ServerPort int
	Secret     string
}

// APIService struct represents the API service. It contains the context, cancel
// function, wait group, configuration, notification queue, API handler and
// HTTP server. The context is used to manage the lifecycle of the service,
// while the wait group is used to wait for background processes to finish.
// The notification queue is used to send notifications, and the API handler
// is used to handle incoming requests.
type APIService struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *Config
	nq     *notification.Queue
	*xhttp.Handler
	*http.Server
}

// New function creates a new api service instance. It takes a context, a config
// struct, and a notification queue as parameters. It returns a pointer to the
// service instance and an error if something goes wrong during the process.
// The function is responsible for setting up the service and its dependencies.
// It handles the configuration, rate limiting, and HTTP server setup. The
// function is designed to be used as a constructor for the service and is
// responsible for initializing all the necessary components for the service
// to function properly.
func New(ctx context.Context, cfg *Config, nq *notification.Queue) (*APIService, error) {
	internalCtx, cancel := context.WithCancel(ctx)
	// create the service
	srv := &APIService{
		ctx:     internalCtx,
		cancel:  cancel,
		cfg:     cfg,
		nq:      nq,
		Handler: xhttp.NewHandler(true),
	}
	// create the rate limiter
	rl := xhttp.NewRateLimiter(internalCtx, 100, time.Minute)
	// register the routes and handlers
	if err := srv.Post(AppsPath, rl.Middleware(srv.generateAppIDHandler)); err != nil {
		return nil, err
	}
	if err := srv.Post(TokensPath, rl.Middleware(srv.requestTokenHandler)); err != nil {
		return nil, err
	}
	if err := srv.Put(TokensPath, rl.Middleware(srv.verifyTokenHandler)); err != nil {
		return nil, err
	}
	// do not use rate limiter middleware for health check handler
	if err := srv.Get(HealthCheckPath, srv.healthCheckHandler); err != nil {
		return nil, err
	}
	// build the http server
	srv.Server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server, cfg.ServerPort),
		Handler: srv.Handler,
	}
	return srv, nil
}

// Start method starts the service by starting the http server.
func (s *APIService) Start() error {
	// start the api server
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop method stops the service. It cancels the context and waits for the
// background processes to finish. It also closes the http server.
func (s *APIService) Stop() error {
	// cancel the context and wait for the background processes finish
	s.cancel()
	if err := s.Close(); err != nil {
		return err
	}
	return nil
}

// Ping method checks if the service is up and running. It sends a GET request
// to the health check endpoint and returns true if the response status code
// is 200 OK, otherwise it returns false. If something goes wrong during the
// process, it returns false.
func (s *APIService) Ping() bool {
	url := fmt.Sprintf("http://%s:%d%s", s.cfg.Server, s.cfg.ServerPort, HealthCheckPath)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false
	}
	defer func() {
		_ = response.Body.Close()
	}()
	return response.StatusCode == http.StatusOK
}
