package authapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/lucasmenendez/apihandler"
	"github.com/simpleauthlink/authapi/db"
	"github.com/simpleauthlink/authapi/email"
)

// Config struct represents the configuration needed to init the service. It
// includes the email configuration, the server hostname, the server port, the
// data path to store the database, and the cleaner cooldown to clean the
// expired tokens.
type Config struct {
	email.EmailConfig
	Server          string
	ServerPort      int
	CleanerCooldown time.Duration
}

// Service struct represents the service that is going to be started. It
// includes the context and the cancel function to stop the service, the wait
// group to wait for the background processes to finish, the configuration,
// the database connection and the api handler.
type Service struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wait       sync.WaitGroup
	cfg        *Config
	db         db.DB
	emailQueue *email.EmailQueue
	handler    *apihandler.Handler
	httpServer *http.Server
}

// New function creates a new service based on the provided context and
// configuration. It initializes the database, creates the service and sets
// the api handlers. If something goes wrong during the process, it returns
// an error.
func New(ctx context.Context, db db.DB, cfg *Config) (*Service, error) {
	internalCtx, cancel := context.WithCancel(ctx)
	// create the service
	srv := &Service{
		ctx:        internalCtx,
		cancel:     cancel,
		cfg:        cfg,
		db:         db,
		emailQueue: email.NewEmailQueue(internalCtx, &cfg.EmailConfig),
		handler:    apihandler.NewHandler(true),
	}
	// set the api handlers
	srv.handler.Post("/user", srv.userTokenHandler)
	srv.handler.Get("/user", srv.validateUserTokenHandler)
	srv.handler.Post("/app", srv.appTokenHandler)
	srv.handler.Put("/app", srv.updateAppHandler)
	srv.handler.Delete("/app", srv.delAppHandler)
	srv.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server, cfg.ServerPort),
		Handler: srv.handler,
	}
	return srv, nil
}

// Start method starts the service. It starts the token cleaner and the api
// server. If something goes wrong during the process, it returns an error.
func (s *Service) Start() error {
	// start the email queue
	s.emailQueue.Start()
	// start the token cleaner in the background
	s.sanityTokenCleaner()
	// start the api server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Stop method stops the service. It cancels the context and waits for the
// background processes to finish. It closes the database. If something goes
// wrong during the process, it returns an error.
func (s *Service) Stop() error {
	// close the database
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("error closing db: %w", err)
	}
	// stop the email queue
	s.emailQueue.Stop()
	// cancel the context and wait for the background processes finish
	s.cancel()
	defer s.wait.Wait()
	return nil
}

// WaitToShutdown method waits for the service to shutdown. It listens for the
// interrupt signal and shutdown the http server and the service. If something
// goes wrong during the process, it returns an error.
func (s *Service) WaitToShutdown() error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer func() {
		if err := s.Stop(); err != nil {
			log.Println(err)
		}
	}()
	return s.httpServer.Shutdown(ctx)
}
