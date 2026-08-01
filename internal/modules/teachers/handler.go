package teachers

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
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create teacher", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Teacher created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.List(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list teachers", err)
		return
	}
	utils.Success(w, http.StatusOK, "Teachers retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Teacher not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get teacher", err)
		return
	}
	utils.Success(w, http.StatusOK, "Teacher retrieved successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Teacher not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update teacher", err)
		return
	}
	utils.Success(w, http.StatusOK, "Teacher updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Teacher not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete teacher", err)
		return
	}
	utils.Success(w, http.StatusOK, "Teacher deleted successfully", nil)
}
