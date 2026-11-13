package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"go-academy/backend/internal/platform/response"
)

// Handler serves administrative user management endpoints.
type Handler struct {
	service *Service
}

// NewHandler constructs a new users Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

// List handles admin user listing with pagination.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	users, total, err := h.service.ListUsers(r.Context(), page, limit)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list users", nil)
		return
	}

	meta := map[string]any{
		"pagination": map[string]any{
			"page":        page,
			"limit":       limit,
			"total_items": total,
		},
	}

	response.JSON(w, http.StatusOK, users, meta)
}

// UpdateStatus handles activating or deactivating a user account.
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid user ID", nil)
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON payload", nil)
		return
	}

	user, err := h.service.UpdateUserStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "User not found", nil)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update user status", nil)
		return
	}

	response.JSON(w, http.StatusOK, user, nil)
}
