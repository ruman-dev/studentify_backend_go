package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"softixa-solutions.com/studentify/internal/modules/auth"
)

func NewRouter(authHandler *auth.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
	})

	return r
}
