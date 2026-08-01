package assignments

import (
	"errors"
	"net/http"
	"strings"

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
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create assignment", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Assignment created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.List(r.Context(), userID, strings.TrimSpace(r.URL.Query().Get("subject_id")))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list assignments", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignments retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Assignment not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get assignment", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignment retrieved successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Assignment not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update assignment", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignment updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Assignment not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete assignment", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignment deleted successfully", nil)
}
