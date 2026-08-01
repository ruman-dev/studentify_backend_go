package profile

import (
	"errors"
	"net/http"

	"softixa-solutions.com/studentify/internal/middleware"
	"softixa-solutions.com/studentify/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.Get(r.Context(), userID)
	if errors.Is(err, utils.ErrUserNotFound) {
		utils.Error(w, http.StatusNotFound, "User not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to load profile", err)
		return
	}

	utils.Success(w, http.StatusOK, "Profile retrieved successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Update(r.Context(), userID, req)
	if errors.Is(err, utils.ErrUserNotFound) {
		utils.Error(w, http.StatusNotFound, "User not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update profile", err)
		return
	}

	utils.Success(w, http.StatusOK, "Profile updated successfully", resp)
}
