package server

import (
	"database/sql"
	"net/http"

	"softixa-solutions.com/studentify/internal/config"
	"softixa-solutions.com/studentify/internal/modules/assignments"
	"softixa-solutions.com/studentify/internal/modules/auth"
	"softixa-solutions.com/studentify/internal/modules/events"
	"softixa-solutions.com/studentify/internal/modules/exams"
	"softixa-solutions.com/studentify/internal/modules/profile"
	"softixa-solutions.com/studentify/internal/modules/subjects"
	"softixa-solutions.com/studentify/internal/modules/teachers"
	"softixa-solutions.com/studentify/internal/route"
	"softixa-solutions.com/studentify/internal/utils"
)

type Server struct {
	cfg    *config.Config
	router http.Handler
}

func New(cfg *config.Config, db *sql.DB) *Server {
	tokens := utils.NewTokenManager(cfg.JWTSecret, cfg.JWTExpiry)

	authHandler := auth.NewHandler(auth.NewService(db, tokens))
	profileHandler := profile.NewHandler(profile.NewService(db))
	teachersHandler := teachers.NewHandler(teachers.NewService(db))
	subjectsHandler := subjects.NewHandler(subjects.NewService(db))
	assignmentsHandler := assignments.NewHandler(assignments.NewService(db))
	examsHandler := exams.NewHandler(exams.NewService(db))
	eventsHandler := events.NewHandler(events.NewService(db))

	return &Server{
		cfg: cfg,
		router: route.NewRouter(route.Dependencies{
			Handlers: route.Handlers{
				Auth:        authHandler,
				Profile:     profileHandler,
				Teachers:    teachersHandler,
				Subjects:    subjectsHandler,
				Assignments: assignmentsHandler,
				Exams:       examsHandler,
				Events:      eventsHandler,
			},
			Tokens: tokens,
		}),
	}
}

func (s *Server) Router() http.Handler {
	return s.router
}

func (s *Server) Addr() string {
	return ":" + s.cfg.Port
}
