package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"softixa-solutions.com/studentify/internal/utils"
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
	if errors.Is(err, utils.ErrInvalidCredentials) {
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

	utils.Success(w, http.StatusCreated, "OTP sent successfully", resp)
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.OTP = strings.TrimSpace(req.OTP)

	var sources []utils.ErrorSource
	if !utils.Email(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if req.OTP == "" {
		sources = append(sources, utils.ErrorSource{Path: "register_otp", Message: "OTP is required"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.VerifyOTP(r.Context(), req)
	if errors.Is(err, utils.ErrUserNotFound) {
		utils.Error(w, http.StatusUnauthorized, "User not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidOTP) {
		utils.Error(w, http.StatusUnauthorized, "Invalid OTP", err)
		return
	}
	if errors.Is(err, utils.ErrOTPExpired) {
		utils.Error(w, http.StatusUnauthorized, "OTP has expired", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "OTP verification failed", err)
		return
	}

	utils.Success(w, http.StatusOK, "OTP verified successfully", resp)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	var sources []utils.ErrorSource
	if !utils.Email(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.ForgotPassword(r.Context(), req)
	if errors.Is(err, utils.ErrUserNotFound) {
		utils.Error(w, http.StatusNotFound, "User not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to send OTP", err)
		return
	}

	utils.Success(w, http.StatusOK, "OTP sent successfully", resp)
}

func (h *Handler) VerifyForgotPasswordOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.OTP = strings.TrimSpace(req.OTP)

	var sources []utils.ErrorSource
	if !utils.Email(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if req.OTP == "" {
		sources = append(sources, utils.ErrorSource{Path: "otp", Message: "OTP is required"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.VerifyForgotPasswordOTP(r.Context(), req)
	if errors.Is(err, utils.ErrUserNotFound) {
		utils.Error(w, http.StatusNotFound, "User not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidOTP) {
		utils.Error(w, http.StatusUnauthorized, "Invalid OTP", err)
		return
	}
	if errors.Is(err, utils.ErrOTPExpired) {
		utils.Error(w, http.StatusUnauthorized, "OTP has expired", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "OTP verification failed", err)
		return
	}

	utils.Success(w, http.StatusOK, "OTP verified successfully", resp)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	resetToken := strings.TrimSpace(r.Header.Get("X-Reset-Token"))
	if resetToken == "" {
		resetToken = strings.TrimSpace(r.Header.Get("Reset-Token"))
	}

	var sources []utils.ErrorSource
	if !utils.Email(req.Email) {
		sources = append(sources, utils.ErrorSource{Path: "email", Message: "Invalid email address"})
	}
	if resetToken == "" {
		sources = append(sources, utils.ErrorSource{Path: "X-Reset-Token", Message: "Reset token header is required"})
	}
	if !utils.Password(req.Password) {
		sources = append(sources, utils.ErrorSource{Path: "password", Message: "Password must be at least 6 characters and include a letter and a number"})
	}
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.ResetPassword(r.Context(), req, resetToken)
	if errors.Is(err, utils.ErrUserNotFound) {
		utils.Error(w, http.StatusNotFound, "User not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidResetToken) {
		utils.Error(w, http.StatusUnauthorized, "Invalid or expired reset token", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Password reset failed", err)
		return
	}

	utils.Success(w, http.StatusOK, "Password updated successfully", resp)
}
