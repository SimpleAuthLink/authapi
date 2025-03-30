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
	"github.com/simpleauthlink/authapi/internal"
	"github.com/simpleauthlink/authapi/notification"
)

type Config struct {
	Server     string
	ServerPort int
	Secret     string
	// demo stuff
	DemoMode     bool
	DemoSMTPAddr string
	DemoSMTPPort int
}

type Service struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wait       sync.WaitGroup
	cfg        *Config
	nq         notification.Queue
	handler    *apihandler.Handler
	httpServer *http.Server
	// demo stuff
	demoMailServer *internal.FakeSMTPServer
	demoMailInbox  chan string
}

func New(ctx context.Context, cfg *Config, nq notification.Queue) (*Service, error) {
	internalCtx, cancel := context.WithCancel(ctx)
	rateLimiter := apihandler.RateLimiter(internalCtx, 1000, 1000, time.Minute*3)
	// create the service
	srv := &Service{
		ctx:     internalCtx,
		cancel:  cancel,
		cfg:     cfg,
		nq:      nq,
		handler: apihandler.NewHandler(true, rateLimiter),
	}
	// demo stuff
	if cfg.DemoMode {
		srv.demoMailInbox = make(chan string, 1)
		srv.demoMailServer = internal.NewFakeSMTPServer(cfg.DemoSMTPAddr,
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

// Start method starts the service.
func (s *Service) Start() error {
	// start the api server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Service) Stop() {
	// cancel the context and wait for the background processes finish
	s.cancel()
	defer s.wait.Wait()
}

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
