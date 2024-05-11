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

type Config struct {
	EmailConfig
	Server          string
	ServerPort      int
	DataPath        string
	CleanerCooldown time.Duration
}

type Service struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wait    sync.WaitGroup
	cfg     *Config
	db      db.DB
	handler *apihandler.Handler
}

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

func (s *Service) Start() error {
	// start the token cleaner
	s.sanityTokenCleaner()
	// start the api server
	addr := fmt.Sprintf("%s:%d", s.cfg.Server, s.cfg.ServerPort)
	return http.ListenAndServe(addr, s.handler)
}

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
