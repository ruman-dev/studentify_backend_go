package server

import (
	"database/sql"
	"net/http"

	"softixa-solutions.com/studentify/internal/config"
	"softixa-solutions.com/studentify/internal/modules/auth"
	"softixa-solutions.com/studentify/internal/route"
)

type Server struct {
	cfg    *config.Config
	router http.Handler
}

func New(cfg *config.Config, db *sql.DB) *Server {
	tokens := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiry)
	repo := auth.NewRepository(db)
	service := auth.NewService(repo, tokens)
	handler := auth.NewHandler(service)

	return &Server{
		cfg:    cfg,
		router: route.NewRouter(handler),
	}
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return ":" + s.cfg.Port
}
