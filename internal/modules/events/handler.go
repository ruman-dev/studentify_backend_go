package events

import (
	"errors"
	"net/http"

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
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req CreateRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Create(r.Context(), userID, req)
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if errors.Is(err, utils.ErrConflict) {
		utils.Error(w, http.StatusBadRequest, "ends_at must be after starts_at", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create event", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Event created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.List(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list events", err)
		return
	}
	utils.Success(w, http.StatusOK, "Events retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

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
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Event not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if errors.Is(err, utils.ErrConflict) {
		utils.Error(w, http.StatusBadRequest, "ends_at must be after starts_at", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update event", err)
		return
	}
	utils.Success(w, http.StatusOK, "Event updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

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
