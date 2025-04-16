package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lucasmenendez/apihandler"
	"github.com/simpleauthlink/authapi/internal/fakesmtpserver"
	"github.com/simpleauthlink/authapi/notification"
)

// Config struct represents the configuration for the API service. It contains
// the server address, server port, and secret key for the service. The server
// address is the address where the service will listen for incoming requests,
// and the server port is the port number where the service will listen for
// incoming requests. The secret key is used to sign and verify tokens. The
// demo mode is used to enable or disable the demo functionality of the service.
// The demo SMTP address and port are used to configure the demo mail server.
// The demo mode is used to enable or disable the demo functionality of the
// service.
type Config struct {
	Server     string
	ServerPort int
	Secret     string
	// demo stuff
	DemoMode     bool
	DemoSMTPAddr string
	DemoSMTPPort int
}

// Service struct represents the API service. It contains the context, cancel
// function, wait group, configuration, notification queue, API handler, HTTP
// server, and demo mail server. The context is used to manage the lifecycle
// of the service, while the wait group is used to wait for background processes
// to finish. The notification queue is used to send notifications, and the API
// handler is used to handle incoming requests. The HTTP server is used to
// serve the API endpoints, and the demo mail server is used to simulate
// sending emails in demo mode.
// The demo mail server is a fake SMTP server that captures emails sent to it
// for testing purposes. The demo mail inbox is a  channel that receives the
// captured emails.
type Service struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wait       sync.WaitGroup
	cfg        *Config
	nq         notification.Queue
	handler    *apihandler.Handler
	httpServer *http.Server
	// demo stuff
	demoMailServer *fakesmtpserver.FakeSMTPServer
	demoMailInbox  chan string
}

// New function creates a new service instance. It takes a context, a config
// struct, and a notification queue as parameters. It returns a pointer to the
// service instance and an error if something goes wrong during the process.
// The function is responsible for setting up the service and its dependencies.
// It handles the configuration, rate limiting, and HTTP server setup. It also
// manages the demo mode functionality, including the demo mail server and
// inbox. The function is designed to be used as a constructor for the service
// and is responsible for initializing all the necessary components for the
// service to function properly.
func New(ctx context.Context, cfg *Config, nq notification.Queue) (*Service, error) {
	internalCtx, cancel := context.WithCancel(ctx)
	// create the service
	srv := &Service{
		ctx:     internalCtx,
		cancel:  cancel,
		cfg:     cfg,
		nq:      nq,
		handler: apihandler.NewHandler(true, nil),
	}
	// demo stuff
	if cfg.DemoMode {
		srv.demoMailInbox = make(chan string, 1)
		srv.demoMailServer = fakesmtpserver.NewServer(cfg.DemoSMTPAddr,
			cfg.DemoSMTPPort, srv.demoMailInbox)
		if err := srv.demoMailServer.Start(internalCtx); err != nil {
			return nil, err
		}
		_ = srv.handler.Get(DemoInboxPath, srv.demoInboxHandler)
	}
	// register the routes and handlers
	_ = srv.handler.Post(AppsPath, srv.generateAppIDHandler)
	_ = srv.handler.Post(TokensPath, srv.requestTokenHandler)
	_ = srv.handler.Put(TokensPath, srv.verifyTokenHandler)
	_ = srv.handler.Get(HealthCheckPath, srv.healthCheckHandler)
	// build the http server
	srv.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server, cfg.ServerPort),
		Handler: srv.handler,
	}
	return srv, nil
}

// Start method starts the service by starting the http server.
func (s *Service) Start() error {
	// start the api server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop method stops the service. It cancels the context and waits for the
// background processes to finish. It also closes the http server and the
// demo mail server if it is running.
func (s *Service) Stop() {
	// cancel the context and wait for the background processes finish
	s.cancel()
	defer s.wait.Wait()
}

// Ping method checks if the service is up and running. It sends a GET request
// to the health check endpoint and returns true if the response status code
// is 200 OK, otherwise it returns false. If something goes wrong during the
// process, it returns false.
func (s *Service) Ping() bool {
	url := fmt.Sprintf("http://%s:%d%s", s.cfg.Server, s.cfg.ServerPort, HealthCheckPath)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false
	}
	return response.StatusCode == http.StatusOK
}

// WaitToShutdown method waits for the service to shutdown. It listens for the
// interrupt signal and shutdown the http server and the service. If something
// goes wrong during the process, it returns an error.
func (s *Service) WaitToShutdown() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer s.Stop()
	return s.httpServer.Shutdown(ctx)
}
