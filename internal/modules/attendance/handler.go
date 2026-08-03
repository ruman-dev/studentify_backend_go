package attendance

import (
	"errors"
	"net/http"
	"time"

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

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.Overview(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get attendance overview", err)
		return
	}
	utils.Success(w, http.StatusOK, "Attendance overview retrieved successfully", resp)
}

func (h *Handler) SubjectDetail(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.SubjectDetail(r.Context(), userID, chi.URLParam(r, "subjectId"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Subject not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get subject attendance", err)
		return
	}
	utils.Success(w, http.StatusOK, "Subject attendance retrieved successfully", resp)
}

func (h *Handler) ListToday(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	day := time.Now()
	if raw := r.URL.Query().Get("date"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, "date must be YYYY-MM-DD", err)
			return
		}
		day = parsed
	}

	resp, err := h.service.ListToday(r.Context(), userID, day)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list today's attendance", err)
		return
	}
	utils.Success(w, http.StatusOK, "Today's attendance retrieved successfully", resp)
}

func (h *Handler) Mark(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req MarkRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Mark(r.Context(), userID, req)
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidInput) {
		utils.Error(w, http.StatusBadRequest, "Invalid session date", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to mark attendance", err)
		return
	}
	utils.Success(w, http.StatusOK, "Attendance marked successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if !utils.DecodeAndValidate(w, r, &req) {
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Attendance record not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update attendance", err)
		return
	}
	utils.Success(w, http.StatusOK, "Attendance updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Attendance record not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete attendance", err)
		return
	}
	utils.Success(w, http.StatusOK, "Attendance deleted successfully", nil)
}
