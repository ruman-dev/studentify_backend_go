package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"softixa-solutions.com/studentify/internal/middleware"
	"softixa-solutions.com/studentify/internal/modules/assignments"
	"softixa-solutions.com/studentify/internal/modules/auth"
	"softixa-solutions.com/studentify/internal/modules/events"
	"softixa-solutions.com/studentify/internal/modules/exams"
	"softixa-solutions.com/studentify/internal/modules/notifications"
	"softixa-solutions.com/studentify/internal/modules/profile"
	"softixa-solutions.com/studentify/internal/modules/subjects"
	"softixa-solutions.com/studentify/internal/modules/teachers"
	"softixa-solutions.com/studentify/internal/utils"
)

// Handlers groups all module HTTP handlers for router wiring.
type Handlers struct {
	Auth          *auth.Handler
	Profile       *profile.Handler
	Teachers      *teachers.Handler
	Subjects      *subjects.Handler
	Assignments   *assignments.Handler
	Exams         *exams.Handler
	Events        *events.Handler
	Notifications *notifications.Handler
}

type Dependencies struct {
	Handlers Handlers
	Tokens   *utils.TokenManager
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.StripSlashes)

	authRoutes(r, deps.Handlers.Auth)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(deps.Tokens))
		protectedRoutes(r, deps.Handlers)
	})

	return r
}

func authRoutes(r chi.Router, h *auth.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/register", h.Register)
		r.Post("/verify-otp", h.VerifyOTP)
		r.Post("/forgot-password", h.ForgotPassword)
		r.Post("/forgot-password/verify-otp", h.VerifyForgotPasswordOTP)
		r.Post("/reset-password", h.ResetPassword)
	})
}

func protectedRoutes(r chi.Router, h Handlers) {
	r.Get("/profile", h.Profile.Get)
	r.Patch("/profile", h.Profile.Update)

	r.Route("/teachers", func(r chi.Router) {
		r.Get("/", h.Teachers.List)
		r.Post("/", h.Teachers.Create)
		r.Get("/{id}", h.Teachers.Get)
		r.Put("/{id}", h.Teachers.Update)
		r.Delete("/{id}", h.Teachers.Delete)
	})

	r.Route("/subjects", func(r chi.Router) {
		r.Get("/", h.Subjects.List)
		r.Post("/", h.Subjects.Create)
		r.Get("/{id}", h.Subjects.Get)
		r.Put("/{id}", h.Subjects.Update)
		r.Delete("/{id}", h.Subjects.Delete)
	})

	r.Route("/assignments", func(r chi.Router) {
		r.Get("/", h.Assignments.List)
		r.Post("/", h.Assignments.Create)
		r.Get("/{id}", h.Assignments.Get)
		r.Put("/{id}", h.Assignments.Update)
		r.Delete("/{id}", h.Assignments.Delete)
	})

	r.Route("/exams", func(r chi.Router) {
		r.Get("/", h.Exams.List)
		r.Get("/types", h.Exams.ListTypes)
		r.Post("/", h.Exams.Create)
		r.Get("/{id}", h.Exams.Get)
		r.Put("/{id}", h.Exams.Update)
		r.Delete("/{id}", h.Exams.Delete)
	})

	r.Route("/events", func(r chi.Router) {
		r.Get("/", h.Events.List)
		r.Post("/", h.Events.Create)
		r.Get("/{id}", h.Events.Get)
		r.Put("/{id}", h.Events.Update)
		r.Delete("/{id}", h.Events.Delete)
	})

	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.Notifications.List)
		r.Post("/", h.Notifications.Create)
		r.Post("/read-all", h.Notifications.MarkAllRead)
		r.Get("/{id}", h.Notifications.Get)
		r.Delete("/{id}", h.Notifications.Delete)
	})
}
