package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"softixa-solutions.com/studentify/internal/handlers"
)

func Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Post("/auth/login", handlers.LoginHandler)
	r.Post("/auth/register", handlers.RegisterHandler)
	return r
}
