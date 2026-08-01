package auth

import (
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
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)

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
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}
	req.FullName = strings.TrimSpace(req.FullName)
	req.Email = strings.TrimSpace(req.Email)
	req.Phone = strings.TrimSpace(req.Phone)

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Registration failed", err)
		return
	}

	utils.Success(w, http.StatusCreated, "OTP sent successfully", resp)
}

func (h *Handler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.OTP = strings.TrimSpace(req.OTP)

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
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)

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
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.OTP = strings.TrimSpace(req.OTP)

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
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	resetToken := strings.TrimSpace(r.Header.Get("X-Reset-Token"))
	if resetToken == "" {
		resetToken = strings.TrimSpace(r.Header.Get("Reset-Token"))
	}
	if resetToken == "" {
		utils.ValidationError(w, []utils.ErrorSource{
			{Path: "X-Reset-Token", Message: "Reset token header is required"},
		})
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
