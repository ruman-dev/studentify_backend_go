package notifications

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
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create notification", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Notification created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var unreadOnly *bool
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get("is_read"))) {
	case "false", "0", "unread":
		v := false
		unreadOnly = &v
	case "true", "1", "read":
		v := true
		unreadOnly = &v
	}

	resp, err := h.service.List(r.Context(), userID, unreadOnly)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list notifications", err)
		return
	}
	utils.Success(w, http.StatusOK, "Notifications retrieved successfully", resp)
}

// Get returns a notification and marks it as read.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.GetAndMarkRead(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Notification not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get notification", err)
		return
	}
	utils.Success(w, http.StatusOK, "Notification retrieved successfully", resp)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	n, err := h.service.MarkAllRead(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to mark notifications as read", err)
		return
	}
	utils.Success(w, http.StatusOK, "Notifications marked as read", map[string]int{"updated": n})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Notification not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete notification", err)
		return
	}
	utils.Success(w, http.StatusOK, "Notification deleted successfully", nil)
}
