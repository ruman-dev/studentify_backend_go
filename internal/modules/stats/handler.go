package stats

import (
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

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	stats, err := h.service.GetOverview(r.Context(), userID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "failed to fetch stats", err)
		return
	}

	utils.Success(w, http.StatusOK, "stats retrieved successfully", stats)
}
