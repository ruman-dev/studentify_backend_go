package events

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"softixa-solutions.com/studentify/internal/middleware"
	"softixa-solutions.com/studentify/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	startsAt, endsAt, sources := validateEventInput(req.Title, req.StartsAt, req.EndsAt, req.EventType)
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Create(r.Context(), userID, req, startsAt, endsAt)
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create event", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Event created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	resp, err := h.service.List(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list events", err)
		return
	}
	utils.Success(w, http.StatusOK, "Events retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	resp, err := h.service.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Event not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get event", err)
		return
	}
	utils.Success(w, http.StatusOK, "Event retrieved successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	startsAt, endsAt, sources := validateEventInput(req.Title, req.StartsAt, req.EndsAt, req.EventType)
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req, startsAt, endsAt)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Event not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update event", err)
		return
	}
	utils.Success(w, http.StatusOK, "Event updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Event not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete event", err)
		return
	}
	utils.Success(w, http.StatusOK, "Event deleted successfully", nil)
}

func validateEventInput(title, startsAtRaw string, endsAtRaw *string, eventType string) (time.Time, *time.Time, []utils.ErrorSource) {
	var sources []utils.ErrorSource
	var startsAt time.Time
	var endsAt *time.Time

	if strings.TrimSpace(title) == "" {
		sources = append(sources, utils.ErrorSource{Path: "title", Message: "Title is required"})
	}
	if strings.TrimSpace(startsAtRaw) == "" {
		sources = append(sources, utils.ErrorSource{Path: "starts_at", Message: "Start time is required"})
	} else {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(startsAtRaw))
		if err != nil {
			sources = append(sources, utils.ErrorSource{Path: "starts_at", Message: "starts_at must be RFC3339 datetime"})
		} else {
			startsAt = parsed
		}
	}
	if endsAtRaw != nil && strings.TrimSpace(*endsAtRaw) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*endsAtRaw))
		if err != nil {
			sources = append(sources, utils.ErrorSource{Path: "ends_at", Message: "ends_at must be RFC3339 datetime"})
		} else {
			endsAt = &parsed
			if !startsAt.IsZero() && parsed.Before(startsAt) {
				sources = append(sources, utils.ErrorSource{Path: "ends_at", Message: "ends_at must be after starts_at"})
			}
		}
	}
	if !IsValidEventType(eventType) {
		sources = append(sources, utils.ErrorSource{Path: "event_type", Message: "event_type must be class, meeting, deadline, or other"})
	}
	return startsAt, endsAt, sources
}
