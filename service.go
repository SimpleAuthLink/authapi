package authapi

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lucasmenendez/apihandler"
	"github.com/lucasmenendez/authapi/db"
	"github.com/lucasmenendez/authapi/db/badger"
)

// Config struct represents the configuration needed to init the service. It
// includes the email configuration, the server hostname, the server port, the
// data path to store the database, and the cleaner cooldown to clean the
// expired tokens.
type Config struct {
	EmailConfig
	Server          string
	ServerPort      int
	DataPath        string
	CleanerCooldown time.Duration
}

// Service struct represents the service that is going to be started. It
// includes the context and the cancel function to stop the service, the wait
// group to wait for the background processes to finish, the configuration,
// the database connection and the api handler.
type Service struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wait    sync.WaitGroup
	cfg     *Config
	db      db.DB
	handler *apihandler.Handler
}

// New function creates a new service based on the provided context and
// configuration. It initializes the database, creates the service and sets
// the api handlers. If something goes wrong during the process, it returns
// an error.
func New(ctx context.Context, cfg *Config) (*Service, error) {
	// init the database with badger driver
	db := new(badger.BadgerDriver)
	if err := db.Init(cfg.DataPath); err != nil {
		return nil, fmt.Errorf("error initializing db: %w", err)
	}
	internalCtx, cancel := context.WithCancel(ctx)
	// create the service
	srv := &Service{
		ctx:     internalCtx,
		cancel:  cancel,
		cfg:     cfg,
		db:      db,
		handler: apihandler.NewHandler(true),
	}
	// set the api handlers
	srv.handler.Post("/user", srv.userTokenHandler)
	srv.handler.Get("/user", srv.validateUserTokenHandler)
	srv.handler.Post("/app", srv.appTokenHandler)
	return srv, nil
}

// Start method starts the service. It starts the token cleaner and the api
// server. If something goes wrong during the process, it returns an error.
func (s *Service) Start() error {
	// start the token cleaner in the background
	s.sanityTokenCleaner()
	// start the api server
	addr := fmt.Sprintf("%s:%d", s.cfg.Server, s.cfg.ServerPort)
	return http.ListenAndServe(addr, s.handler)
}

// Stop method stops the service. It cancels the context and waits for the
// background processes to finish. It closes the database. If something goes
// wrong during the process, it returns an error.
func (s *Service) Stop() error {
	// cancel the context and wait for the background processes finish
	s.cancel()
	defer s.wait.Wait()
	// close the database
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("error closing db: %w", err)
	}
	return nil
}
