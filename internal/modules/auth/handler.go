package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"softixa-solutions.com/studentify/internal/utils"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	var sources []utils.ErrorSource
	if !utils.Email(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if req.Password == "" {
		sources = append(sources, utils.ErrorSource{Path: "password", Message: "Password is required"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if errors.Is(err, ErrInvalidCredentials) {
		utils.Error(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Login failed", err)
		return
	}

	utils.Success(w, http.StatusOK, "Login successful", resp)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	var sources []utils.ErrorSource
	if req.FullName == "" {
		sources = append(sources, utils.ErrorSource{Path: "full_name", Message: "Full name is required"})
	}
	if !utils.Email(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if !utils.Phone(req.Phone) {
		sources = append(sources, utils.ErrorSource{Path: "phone", Message: "Invalid phone number. Use international format e.g. +1234567890"})
	}
	if !utils.Password(req.Password) {
		sources = append(sources, utils.ErrorSource{Path: "password", Message: "Password must be at least 6 characters and include a letter and a number"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Registration failed", err)
		return
	}

	utils.Success(w, http.StatusCreated, "Registration successful", resp)
}
