package api

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
	"github.com/simpleauthlink/authapi/helpers"
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

// New function creates a new service based on the provided context, the db
// interface and configuration. It initializes the email queue, creates the
// service and sets the api handlers. If something goes wrong during the
// process, it returns an error.
func New(ctx context.Context, db db.DB, cfg *Config) (*Service, error) {
	internalCtx, cancel := context.WithCancel(ctx)
	emailQueue, err := email.NewEmailQueue(internalCtx, &cfg.EmailConfig)
	if err != nil {
		if emailQueue == nil {
			cancel()
			return nil, err
		}
		log.Println("WRN: something occurs during email queue creation:", err)
	}
	// create the service
	srv := &Service{
		ctx:        internalCtx,
		cancel:     cancel,
		cfg:        cfg,
		db:         db,
		emailQueue: emailQueue,
		handler: apihandler.NewHandler(&apihandler.Config{
			CORS: true,
			RateLimitConfig: &apihandler.RateLimitConfig{
				Rate:  2,
				Limit: 10,
			},
		}),
	}
	srv.handler.Get(helpers.HealthCheckPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// user handlers
	srv.handler.Post(helpers.UserEndpointPath, srv.userTokenHandler)
	srv.handler.Get(helpers.UserEndpointPath, srv.validateUserTokenHandler)
	// app handlers
	srv.handler.Get(helpers.AppEndpointPath, srv.appHandler)
	srv.handler.Post(helpers.AppEndpointPath, srv.appTokenHandler)
	srv.handler.Put(helpers.AppEndpointPath, srv.updateAppHandler)
	srv.handler.Delete(helpers.AppEndpointPath, srv.delAppHandler)
	// build the http server
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
