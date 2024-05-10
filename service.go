package authapi

import (
	"fmt"
	"net/http"

	"github.com/lucasmenendez/apihandler"
	"github.com/lucasmenendez/authapi/db"
	"github.com/lucasmenendez/authapi/db/badger"
)

type Config struct {
	EmailConfig
	Server     string
	ServerPort int
	DataPath   string
}

type Service struct {
	cfg     *Config
	db      db.DB
	handler *apihandler.Handler
}

func New(cfg *Config) (*Service, error) {
	// init the database with badger driver
	db := new(badger.BadgerDriver)
	if err := db.Init(cfg.DataPath); err != nil {
		return nil, fmt.Errorf("error initializing db: %w", err)
	}
	// create the service
	srv := &Service{
		cfg:     cfg,
		db:      db,
		handler: apihandler.NewHandler(true),
	}
	// set the api handlers
	srv.handler.Post("", srv.userTokenHandler)
	srv.handler.Get("", srv.validateUserTokenHandler)
	srv.handler.Post("/app", srv.appTokenHandler)
	return srv, nil
}

func (s *Service) Start() error {
	// start the api server
	addr := fmt.Sprintf("%s:%d", s.cfg.Server, s.cfg.ServerPort)
	return http.ListenAndServe(addr, s.handler)
}

func (s *Service) Stop() error {
	// close the database
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("error closing db: %w", err)
	}
	return nil
}
