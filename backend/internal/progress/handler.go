package progress

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/auth"
	"go-academy/backend/internal/platform/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type updateProgressRequest struct {
	Status string `json:"status"` // started, completed
}

// GetProgress handles GET /api/v1/lessons/{id}/progress.
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user claim", nil)
		return
	}

	lessonIDStr := chi.URLParam(r, "id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	prog, err := h.service.GetLessonProgress(r.Context(), userID, lessonID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "No progress found for lesson", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get progress", nil)
		return
	}

	response.JSON(w, http.StatusOK, prog, nil)
}

// UpdateProgress handles PUT /api/v1/lessons/{id}/progress.
func (h *Handler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaims(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid user claim", nil)
		return
	}

	lessonIDStr := chi.URLParam(r, "id")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid lesson ID", nil)
		return
	}

	var req updateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Malformed request body", nil)
		return
	}

	prog, err := h.service.UpdateLessonProgress(r.Context(), userID, lessonID, req.Status)
	if err != nil {
		if errors.Is(err, ErrPrerequisitesUnmet) {
			response.Error(w, http.StatusBadRequest, "PREREQUISITES_NOT_MET", "Cannot complete lesson: prerequisite lessons must be completed first", nil)
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error(), nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update progress", nil)
		return
	}

	response.JSON(w, http.StatusOK, prog, nil)
}
