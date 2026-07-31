package exams

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	examDate, sources := validateExamInput(req.SubjectID, req.Title, req.ExamDate)
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Create(r.Context(), userID, req, examDate)
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create exam", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Exam created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.List(r.Context(), userID, strings.TrimSpace(r.URL.Query().Get("subject_id")))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list exams", err)
		return
	}
	utils.Success(w, http.StatusOK, "Exams retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Exam not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get exam", err)
		return
	}
	utils.Success(w, http.StatusOK, "Exam retrieved successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	examDate, sources := validateExamInput(req.SubjectID, req.Title, req.ExamDate)
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req, examDate)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Exam not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update exam", err)
		return
	}
	utils.Success(w, http.StatusOK, "Exam updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Exam not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete exam", err)
		return
	}
	utils.Success(w, http.StatusOK, "Exam deleted successfully", nil)
}

func validateExamInput(subjectID, title, examDateRaw string) (time.Time, []utils.ErrorSource) {
	var sources []utils.ErrorSource
	var examDate time.Time

	if strings.TrimSpace(subjectID) == "" {
		sources = append(sources, utils.ErrorSource{Path: "subject_id", Message: "Subject is required"})
	}
	if strings.TrimSpace(title) == "" {
		sources = append(sources, utils.ErrorSource{Path: "title", Message: "Title is required"})
	}
	if strings.TrimSpace(examDateRaw) == "" {
		sources = append(sources, utils.ErrorSource{Path: "exam_date", Message: "Exam date is required"})
	} else {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(examDateRaw))
		if err != nil {
			sources = append(sources, utils.ErrorSource{Path: "exam_date", Message: "Exam date must be RFC3339 datetime"})
		} else {
			examDate = parsed
		}
	}
	return examDate, sources
}
