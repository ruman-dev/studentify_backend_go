package profile

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	var sources []utils.ErrorSource
	if req.FullName != nil && strings.TrimSpace(*req.FullName) == "" {
		sources = append(sources, utils.ErrorSource{Path: "full_name", Message: "Full name cannot be empty"})
	}
	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		if phone != "" && !utils.Phone(phone) {
			sources = append(sources, utils.ErrorSource{Path: "phone", Message: "Invalid phone number. Use international format e.g. +1234567890"})
		}
	}
	if req.DateOfBirth != nil {
		raw := strings.TrimSpace(*req.DateOfBirth)
		if raw != "" {
			if _, err := time.Parse("2006-01-02", raw); err != nil {
				sources = append(sources, utils.ErrorSource{Path: "date_of_birth", Message: "Date of birth must be YYYY-MM-DD"})
			}
		}
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
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
