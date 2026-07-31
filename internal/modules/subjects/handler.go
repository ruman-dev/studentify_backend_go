package subjects

import (
	"encoding/json"
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

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		utils.ValidationError(w, []utils.ErrorSource{{Path: "name", Message: "Name is required"}})
		return
	}

	resp, err := h.service.Create(r.Context(), userID, req)
	if errors.Is(err, utils.ErrInvalidTeacher) {
		utils.Error(w, http.StatusBadRequest, "Teacher not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create subject", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Subject created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	resp, err := h.service.List(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list subjects", err)
		return
	}
	utils.Success(w, http.StatusOK, "Subjects retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	resp, err := h.service.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Subject not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get subject", err)
		return
	}
	utils.Success(w, http.StatusOK, "Subject retrieved successfully", resp)
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

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		utils.ValidationError(w, []utils.ErrorSource{{Path: "name", Message: "Name is required"}})
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Subject not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidTeacher) {
		utils.Error(w, http.StatusBadRequest, "Teacher not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update subject", err)
		return
	}
	utils.Success(w, http.StatusOK, "Subject updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", utils.ErrInvalidToken)
		return
	}

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Subject not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete subject", err)
		return
	}
	utils.Success(w, http.StatusOK, "Subject deleted successfully", nil)
}
