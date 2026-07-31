package assignments

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

	dueDate, sources := validateAssignmentInput(req.SubjectID, req.Title, req.DueDate, req.Status)
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Create(r.Context(), userID, req, dueDate)
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to create assignment", err)
		return
	}
	utils.Success(w, http.StatusCreated, "Assignment created successfully", resp)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.List(r.Context(), userID, strings.TrimSpace(r.URL.Query().Get("subject_id")))
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to list assignments", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignments retrieved successfully", resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	resp, err := h.service.Get(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Assignment not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to get assignment", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignment retrieved successfully", resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return
	}

	dueDate, sources := validateAssignmentInput(req.SubjectID, req.Title, req.DueDate, req.Status)
	if len(sources) > 0 {
		utils.ValidationError(w, sources)
		return
	}

	resp, err := h.service.Update(r.Context(), userID, chi.URLParam(r, "id"), req, dueDate)
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Assignment not found", err)
		return
	}
	if errors.Is(err, utils.ErrInvalidSubject) {
		utils.Error(w, http.StatusBadRequest, "Subject not found for this user", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to update assignment", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignment updated successfully", resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.service.Delete(r.Context(), userID, chi.URLParam(r, "id"))
	if errors.Is(err, utils.ErrNotFound) {
		utils.Error(w, http.StatusNotFound, "Assignment not found", err)
		return
	}
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "Failed to delete assignment", err)
		return
	}
	utils.Success(w, http.StatusOK, "Assignment deleted successfully", nil)
}

func validateAssignmentInput(subjectID, title, dueDateRaw, status string) (time.Time, []utils.ErrorSource) {
	var sources []utils.ErrorSource
	var dueDate time.Time

	if strings.TrimSpace(subjectID) == "" {
		sources = append(sources, utils.ErrorSource{Path: "subject_id", Message: "Subject is required"})
	}
	if strings.TrimSpace(title) == "" {
		sources = append(sources, utils.ErrorSource{Path: "title", Message: "Title is required"})
	}
	if strings.TrimSpace(dueDateRaw) == "" {
		sources = append(sources, utils.ErrorSource{Path: "due_date", Message: "Due date is required"})
	} else {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(dueDateRaw))
		if err != nil {
			sources = append(sources, utils.ErrorSource{Path: "due_date", Message: "Due date must be RFC3339 datetime"})
		} else {
			dueDate = parsed
		}
	}
	if !IsValidStatus(status) {
		sources = append(sources, utils.ErrorSource{Path: "status", Message: "Status must be pending, submitted, graded, or overdue"})
	}
	return dueDate, sources
}
